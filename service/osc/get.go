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
	if fileExists(dirPath) {
		return errors.New("directory does not exist")
	}

	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
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
		return err
	}

	objList, err := osc.osfs.ObjectList()
	if err != nil {
		return err
	}

	downlaodList := getDownloadList(fileList, objList, dirPath)

	jobs := make(chan utils.Object, len(downlaodList))
	resultChan := make(chan error, len(downlaodList))

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

	for err := range resultChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func getDownloadList(fileList, objList []*utils.Object, dirPath string) []*utils.Object {
	downloadList := []*utils.Object{}

	for _, obj := range objList {
		chk := false
		for _, file := range fileList {
			fileName, _ := filepath.Rel(dirPath, file.Key)
			if obj.Key == strings.ReplaceAll(fileName, "\\", "/") {
				chk = true
				if obj.Size != file.Size {
					downloadList = append(downloadList, obj)
				}
				break
			}
		}
		if !chk {
			downloadList = append(downloadList, obj)
		}
	}

	return downloadList
}

func combinePaths(basePath, relativePath string) (string, error) {
	bName := filepath.Base(basePath)

	parts := strings.Split(relativePath, "/")
	if bName == parts[0] {
		return filepath.Join(basePath, strings.Join(parts[1:], "/")), nil
	}
	return filepath.Join(basePath, relativePath), nil
}

func mGetWorker(osc *OSController, dirPath string, jobs chan utils.Object, resultChan chan<- error) {
	for obj := range jobs {
		src, err := osc.osfs.Open(obj.Key)
		if err != nil {
			resultChan <- err
			continue
		}
		defer src.Close()

		fileName, err := combinePaths(dirPath, obj.Key)
		if err != nil {
			resultChan <- err
			continue
		}

		err = os.MkdirAll(filepath.Dir(fileName), 0755)
		if err != nil {
			resultChan <- err
			continue
		}

		fmt.Println(fileName)

		dst, err := os.Create(fileName)
		if err != nil {
			resultChan <- err
			continue
		}
		defer dst.Close()

		n, err := io.Copy(dst, src)
		if err != nil {
			resultChan <- err
			continue
		}

		if n != obj.Size {
			resultChan <- errors.New("get failed")
			continue
		}

		dst.Close()
		src.Close()

		resultChan <- nil
	}
}
