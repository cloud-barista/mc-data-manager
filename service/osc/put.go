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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloud-barista/cm-data-mold/pkg/utils"
)

func (osc *OSController) MPut(dirPath string) error {
	if err := osc.osfs.CreateBucket(); err != nil {
		osc.logWrite("Error", "CreateBucket error", err)
		return err
	}

	if utils.FileExists(dirPath) {
		err := errors.New("directory does not exist")
		osc.logWrite("Error", "FileExists error", err)
		return err
	}

	var objList []utils.Object

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			objList = append(objList, utils.Object{
				ChecksumAlgorithm: []string{},
				ETag:              "",
				Key:               path,
				LastModified:      info.ModTime(),
				Size:              info.Size(),
				StorageClass:      "Standard",
			})
		}

		return nil
	})

	if err != nil {
		osc.logWrite("Error", "Walk error", err)
		return err
	}

	jobs := make(chan utils.Object, len(objList))
	resultChan := make(chan Result, len(objList))

	var wg sync.WaitGroup
	for i := 0; i < osc.threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mPutWorker(osc, dirPath, jobs, resultChan)
		}()
	}

	for _, obj := range objList {
		jobs <- obj
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for ret := range resultChan {
		if ret.err != nil {
			osc.logWrite("Error", fmt.Sprintf("Import failed: %s", ret.name), ret.err)
		}
	}
	return nil
}

func mPutWorker(osc *OSController, dirPath string, jobs chan utils.Object, resultChan chan<- Result) {
	for obj := range jobs {
		ret := Result{
			name: obj.Key,
			err:  nil,
		}

		src, err := os.Open(obj.Key)
		if err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}
		defer src.Close()

		fileName, err := filepath.Rel(dirPath, obj.Key)
		if err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}
		fileName = strings.ReplaceAll(filepath.Join(filepath.Base(dirPath), fileName), "\\", "/")

		dst, err := osc.osfs.Create(fileName)
		if err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}
		defer dst.Close()

		n, err := io.Copy(dst, src)
		if err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}

		if n != obj.Size {
			ret.err = errors.New("put failed")
			resultChan <- ret
			continue
		}

		dst.Close()
		src.Close()

		osc.logWrite("Info", fmt.Sprintf("Import success: %s -> %s", obj.Key, fileName), nil)

		resultChan <- ret
	}
}
