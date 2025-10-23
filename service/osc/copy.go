/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package osc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/filtering"
)

func (src *OSController) Copy(dst *OSController, flt *filtering.ObjectFilter) error {
	if err := dst.osfs.CreateBucket(); err != nil {
		src.logWrite("Error", "CreateBucket error", err)
		return err
	}

	srcObjList, err := src.ObjectListWithFilter(flt)
	if err != nil {
		src.logWrite("Error", "source objectList error", err)
		return err
	}

	if b, err := json.MarshalIndent(srcObjList, "", "  "); err == nil {
		fmt.Println("Filtered Objects:", string(b))
	}

	dstObjList, err := dst.osfs.ObjectList()
	if err != nil {
		src.logWrite("Error", "target objectList error", err)
		return err
	}

	path := ""
	if flt != nil && flt.Path != "" {
		path = strings.TrimPrefix(flt.Path, "/")
	}

	copyList, skipList := getDownloadList(dstObjList, srcObjList, path, flt.PathExcludeYn)

	for _, skip := range skipList {
		src.logWrite("Info", fmt.Sprintf("skip file : %s", skip.Key), nil)
	}

	jobs := make(chan models.Object, len(copyList))
	resultChan := make(chan Result, len(copyList))

	var wg sync.WaitGroup
	for i := 0; i < src.threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			copyWorker(src, dst, jobs, resultChan)
		}()
	}

	for _, obj := range copyList {
		jobs <- *obj
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for ret := range resultChan {
		if ret.err != nil {
			src.logWrite("Error", fmt.Sprintf("Migration failed: %s", ret.name), ret.err)
		}
	}

	return nil
}

func copyWorker(src *OSController, dst *OSController, jobs chan models.Object, resultChan chan<- Result) {
	for obj := range jobs {
		ret := Result{
			name: obj.Key,
			err:  nil,
		}

		srcFile, err := src.osfs.Open(obj.Key)
		if err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}

		dstFile, err := dst.osfs.Create(obj.Key)
		if err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}

		n, err := io.Copy(dstFile, srcFile)
		if err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}

		if n != obj.Size {
			ret.err = errors.New("copy failed")
			resultChan <- ret
			continue
		}

		if err := srcFile.Close(); err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}

		if err := dstFile.Close(); err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}

		src.logWrite("Info", fmt.Sprintf("Migration success: src:/%s -> dst:/%s", obj.Key, obj.Key), nil)

		resultChan <- ret
	}
}
