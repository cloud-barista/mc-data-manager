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

	var fileList []*utils.Object

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileList = append(fileList, &utils.Object{
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

	jobs := make(chan utils.Object, len(downlaodList))
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

func getDownloadList(fileList, objList []*utils.Object, dirPath string) ([]*utils.Object, []*utils.Object) {
	downloadList := []*utils.Object{}
	skipList := []*utils.Object{}

	for _, obj := range objList {
		chk := false
		for _, file := range fileList {
			fileName, _ := filepath.Rel(dirPath, file.Key)
			objName, _ := filepath.Rel(filepath.Base(dirPath), obj.Key)
			if objName == fileName {
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

func mGetWorker(osc *OSController, dirPath string, jobs chan utils.Object, resultChan chan<- Result) {
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
