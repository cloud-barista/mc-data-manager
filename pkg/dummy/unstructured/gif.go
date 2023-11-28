package unstructured

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/sirupsen/logrus"
)

// GIF generation function using gofakeit
//
// CapacitySize is in GB and generates gif files
// within the entered dummyDir path.
func GenerateRandomGIF(dummyDir string, capacitySize int) error {
	dummyDir = filepath.Join(dummyDir, "gif")
	if err := utils.IsDir(dummyDir); err != nil {
		logrus.Errorf("IsDir function error : %v", err)
		return err
	}

	tempPath := filepath.Join(dummyDir, "tmpImg")
	if err := os.MkdirAll(tempPath, 0755); err != nil {
		logrus.Errorf("MkdirAll function error : %v", err)
		return err
	}
	defer os.RemoveAll(tempPath)

	logrus.Info("start png generation")
	if err := GenerateRandomPNGImage(tempPath, 1); err != nil {
		logrus.Error("failed to generate png")
		return err
	}
	logrus.Info("successfully generated png")

	var files []string
	size := capacitySize * 34 * 10

	err := filepath.Walk(tempPath, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".png" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		logrus.Errorf("Walk function error : %v", err)
		return err
	}

	var imgList []image.Image
	for _, imgName := range files {
		imgFile, err := os.Open(imgName)
		if err != nil {
			logrus.Errorf("file open error : %v", err)
			return err
		}
		defer imgFile.Close()

		img, err := png.Decode(imgFile)
		if err != nil {
			logrus.Errorf("file decoding error : %v", err)
			return err
		}
		imgList = append(imgList, img)
	}

	countNum := make(chan int, size)
	resultChan := make(chan error, size)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomGIFWorker(imgList, countNum, dummyDir, resultChan)
		}()
	}

	for i := 0; i < size; i++ {
		countNum <- i
	}
	close(countNum)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for err := range resultChan {
		if err != nil {
			logrus.Errorf("result error : %v", err)
			return err
		}
	}

	return nil
}

// gif worker
func randomGIFWorker(imgList []image.Image, countNum chan int, tmpDir string, resultChan chan<- error) {
	for cnt := range countNum {
		randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

		randGen.Shuffle(len(imgList), func(i, j int) {
			imgList[i], imgList[j] = imgList[j], imgList[i]
		})

		delay := 10
		gifImage := &gif.GIF{}

		for i, img := range imgList {
			if i == 10 {
				break
			}
			bounds := img.Bounds()
			palettedImage := image.NewPaletted(bounds, palette.Plan9)
			draw.FloydSteinberg.Draw(palettedImage, bounds, img, image.Point{})

			gifImage.Image = append(gifImage.Image, palettedImage)
			gifImage.Delay = append(gifImage.Delay, delay)
		}

		gifFile, err := os.Create(fmt.Sprintf("%s/randomGIF_%d.gif", tmpDir, cnt))
		if err != nil {
			resultChan <- err
		}
		defer gifFile.Close()

		err = gif.EncodeAll(gifFile, gifImage)
		if err == nil {
			logrus.Infof("Creation success: %v", gifFile.Name())
		}

		if cerr := gifFile.Close(); cerr != nil {
			err = cerr
		}
		resultChan <- err
	}
}
