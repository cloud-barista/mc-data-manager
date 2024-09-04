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

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
)

func (osc *OSController) MGet(dirPath string) error {
	if utils.FileExists(dirPath) {
		err := errors.New("directory does not exist")
		osc.logWrite("Error", "FileExists error", err)
		return err
	}

	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		osc.logWrite("Error", "MkdirAll error", err)
		return err
	}

	var fileList []*models.Object

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileList = append(fileList, &models.Object{
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

	objList, err := osc.osfs.ObjectList()
	if err != nil {
		osc.logWrite("Error", "ObjectList error", err)
		return err
	}

	downlaodList, skipList := getDownloadList(fileList, objList, dirPath)

	for _, skip := range skipList {
		osc.logWrite("Info", fmt.Sprintf("skip file : %s", skip.Key), nil)
	}

	jobs := make(chan models.Object, len(downlaodList))
	resultChan := make(chan Result, len(downlaodList))

	var wg sync.WaitGroup
	for i := 0; i < osc.threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mGetWorker(osc, dirPath, jobs, resultChan)
		}()
	}

	for _, obj := range downlaodList {
		jobs <- *obj
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for ret := range resultChan {
		if ret.err != nil {
			osc.logWrite("Error", fmt.Sprintf("Export failed: %s", ret.name), ret.err)
		}
	}
	return nil
}

func getDownloadList(fileList, objList []*models.Object, dirPath string) ([]*models.Object, []*models.Object) {
	downloadList := []*models.Object{}
	skipList := []*models.Object{}

	for _, obj := range objList {
		chk := false
		for _, file := range fileList {
			fileName, _ := filepath.Rel(dirPath, file.Key)
			objName, _ := filepath.Rel(filepath.Base(dirPath), obj.Key)
			if strings.Contains(objName, fileName) {
				chk = true
				if obj.Size != file.Size {
					downloadList = append(downloadList, obj)
				} else {
					skipList = append(skipList, obj)
				}
				break
			}
		}
		if !chk {
			downloadList = append(downloadList, obj)
		}
	}

	return downloadList, skipList
}

func combinePaths(basePath, relativePath string) (string, error) {
	bName := filepath.Base(basePath)

	parts := strings.Split(relativePath, "/")
	if bName == parts[0] {
		return filepath.Join(basePath, strings.Join(parts[1:], "/")), nil
	}
	return filepath.Join(basePath, relativePath), nil
}

func mGetWorker(osc *OSController, dirPath string, jobs chan models.Object, resultChan chan<- Result) {
	for obj := range jobs {
		ret := Result{
			name: obj.Key,
			err:  nil,
		}

		src, err := osc.osfs.Open(obj.Key)
		if err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}
		defer src.Close()

		fileName, err := combinePaths(dirPath, obj.Key)
		if err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}

		err = os.MkdirAll(filepath.Dir(fileName), 0755)
		if err != nil {
			ret.err = err
			resultChan <- ret
			continue
		}

		dst, err := os.Create(fileName)
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
			ret.err = errors.New("get failed")
			resultChan <- ret
			continue
		}

		dst.Close()
		src.Close()

		osc.logWrite("Info", fmt.Sprintf("Export success: %s -> %s", obj.Key, fileName), nil)

		resultChan <- ret
	}
}
