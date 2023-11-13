package osc

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/cloud-barista/cm-data-mold/pkg/utils"
)

func (src *OSController) Copy(dst *OSController) error {
	if err := dst.osfs.CreateBucket(); err != nil {
		src.logWrite("Error", "CreateBucket error", err)
		return err
	}

	srcObjList, err := src.osfs.ObjectList()
	if err != nil {
		src.logWrite("Error", "source objectList error", err)
		return err
	}

	dstObjList, err := dst.osfs.ObjectList()
	if err != nil {
		src.logWrite("Error", "target objectList error", err)
		return err
	}

	copyList, skipList := getDownloadList(dstObjList, srcObjList, "")

	for _, skip := range skipList {
		src.logWrite("Info", fmt.Sprintf("skip file : %s", skip.Key), nil)
	}

	jobs := make(chan utils.Object, len(copyList))
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

func copyWorker(src *OSController, dst *OSController, jobs chan utils.Object, resultChan chan<- Result) {
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
