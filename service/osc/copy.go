package osc

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/cloud-barista/cm-data-mold/pkg/utils"
)

func (src *OSController) Copy(dst *OSController) error {
	dst.osfs.CreateBucket()

	objList, err := src.osfs.ObjectList()
	if err != nil {
		return nil
	}

	jobs := make(chan utils.Object)
	resultChan := make(chan error)

	var wg sync.WaitGroup
	for i := 0; i < src.threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			copyWorker(src, dst, jobs, resultChan)
		}()
	}

	for _, obj := range objList {
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

func copyWorker(src *OSController, dst *OSController, jobs chan utils.Object, resultChan chan<- error) {
	for obj := range jobs {
		srcFile, err := src.osfs.Open(obj.Key)
		if err != nil {
			resultChan <- err
			continue
		}

		fmt.Println(obj.Key)
		dstFile, err := dst.osfs.Create(obj.Key)
		if err != nil {
			resultChan <- err
			continue
		}

		n, err := io.Copy(dstFile, srcFile)
		if err != nil {
			resultChan <- err
			continue
		}

		if n != obj.Size {
			resultChan <- errors.New("copy failed")
			continue
		}

		if err := srcFile.Close(); err != nil {
			resultChan <- err
			continue
		}

		if err := dstFile.Close(); err != nil {
			resultChan <- err
			continue
		}
	}
}
