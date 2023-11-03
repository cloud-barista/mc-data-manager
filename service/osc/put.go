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

func fileExists(filePath string) bool {
	if fi, err := os.Stat(filePath); os.IsExist(err) {
		return !fi.IsDir()
	}
	return false
}

func (osc *OSController) Put(filePath string) error {
	if !fileExists(filePath) {
		return errors.New("file does not exist")
	}

	src, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := osc.osfs.Create(filepath.Base(filePath))
	if err != nil {
		return err
	}

	n, err := io.Copy(dst, src)
	if err != nil {
		return err
	}

	sinfo, err := src.Stat()
	if err != nil {
		return err
	}

	if n != sinfo.Size() {
		return errors.New("put failed")
	}

	return nil
}

func (osc *OSController) MPut(dirPath string) error {
	osc.osfs.CreateBucket()

	if fileExists(dirPath) {
		return errors.New("directory does not exist")
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
		return err
	}

	jobs := make(chan utils.Object, len(objList))
	resultChan := make(chan error, len(objList))

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

	for err := range resultChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func mPutWorker(osc *OSController, dirPath string, jobs chan utils.Object, resultChan chan<- error) {
	for obj := range jobs {
		src, err := os.Open(obj.Key)
		if err != nil {
			resultChan <- err
			continue
		}
		defer src.Close()

		fileName, err := filepath.Rel(dirPath, obj.Key)
		if err != nil {
			resultChan <- err
			continue
		}
		fileName = strings.ReplaceAll(filepath.Join(filepath.Base(dirPath), fileName), "\\", "/")

		fmt.Println(fileName)

		dst, err := osc.osfs.Create(fileName)
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
			resultChan <- errors.New("put failed")
			continue
		}

		dst.Close()
		src.Close()

		resultChan <- nil
	}
}
