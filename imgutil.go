package main

import (
	"bytes"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/valyala/fasthttp"
	"image"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// MkImageFolder safely creates the image folder if it is not cloned by git (since it is empty)
func MkImageFolder() {
	dir := "www/images"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, os.FileMode(0700))

		if err != nil {
			log.Fatalf("- Error making %s - %v", dir, err)
		}
	}
}

// LoadImage returns an image object for fileInput
func LoadImage(fileInput string) (image.Image, error) {
	imgBytes, err := ioutil.ReadFile(fileInput)
	if err != nil {
		log.Fatalf("- Failed loading %s - %s", fileInput, err)
	}

	img, _, err := image.Decode(bytes.NewBuffer(imgBytes))
	if err != nil {
		return nil, err
	}

	return img, nil
}

// MainImageColor calculates the median color of cropped image
func MainImageColor(image string) string {
	img, err := LoadImage(image)
	if nil != err {
		log.Fatalf("- Failed loading image %s - %s", image, err)
		return ""
	}

	cols, err := prominentcolor.KmeansWithArgs(prominentcolor.ArgumentDefault, img)
	if err != nil {
		// The only meaningful error returned here is with all transparent images
		return ""
	}

	col := cols[0].AsString()
	return col
}

// UpdateCurrentImage updates the currentImage, along with it's hash and color
func UpdateCurrentImage() {
	currentImage, currentHash = GetRandomImage()
	colorTmp := MainImageColor("www/images/" + currentImage)
	if len(colorTmp) > 1 {
		currentImageColor = colorTmp
	} else {
		currentImageColor = "000000"
	}
}

// UpdateImageCache opens the www/images directory and caches a list of FileInfos
func UpdateImageCache() []os.FileInfo {
	dir := "www/images"
	dirOpen, _ := os.Open(dir)
	tmpImages, err := dirOpen.Readdir(0)
	if err != nil {
		panic(err)
	}
	return tmpImages
}

// GetRandomImage chooses a random image name from the image cache
func GetRandomImage() (string, string) {
	filename := images[rGen.Intn(len(images))].Name()
	return filename, strings.TrimSuffix(filename, filepath.Ext(filename))
}

// ImageHandler is our own RequestHandler with a CacheDuration of 0
func ImageHandler(root string) fasthttp.RequestHandler {
	fs := &fasthttp.FS{
		Root:               root,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: true,
		AcceptByteRange:    true,
		CacheDuration:      0,
	}
	return fs.NewRequestHandler()
}
