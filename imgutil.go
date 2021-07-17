package main

import (
	"bytes"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/h2non/bimg"
	"github.com/valyala/fasthttp"
	"image"
	_ "image/png" // used by prominent color
	"io/ioutil"
	"log"
	"math"
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

func SaveFinal(path string) error {
	buffer, ext, err := ConvertAndCompress(path)

	compressedPath := path + ext
	removePath := path

	fiOriginal, err := os.Stat(path)
	if err != nil {
		return err
	}

	err = bimg.Write(compressedPath, buffer)
	if err != nil {
		return err
	}

	fiNew, err := os.Stat(compressedPath)
	if err != nil {
		return err
	}

	// We want to do this check in case the original image was more efficiently compressed than ours
	// fiNew is path + ext (compressed), fiOriginal is path
	if fiNew.Size() > fiOriginal.Size() {
		removePath = path + ext
		compressedPath = path
	}

	// Move compressed file to www/images/<file hash>
	hash, err := GetFileHash(compressedPath)
	if err != nil {
		return err
	}
	err = os.Rename(compressedPath, "www/images/"+hash)
	if err != nil {
		return err
	}

	err = os.Remove(removePath)

	return err
}

// ConvertAndCompress will convert the image to jpg if it's non-transparent, and compress
// if it meets the requirements for being compressed
func ConvertAndCompress(path string) ([]byte, string, error) {
	buffer, err := bimg.Read(path)
	if err != nil {
		return nil, "", err
	}

	ext, buffer, err := ConvertImage(buffer)
	if err != nil {
		return nil, "", err
	}

	buffer, err = CompressImage(buffer)
	if err != nil {
		return nil, "", err
	}

	return buffer, ext, nil
}

// ConvertImage will take path and convert the image to a png
func ConvertImage(buffer []byte) (string, []byte, error) {
	ext := ""
	img, err := bimg.NewImage(buffer).Metadata()
	if err != nil {
		return ext, nil, err
	}

	if img.Alpha == true {
		ext = ".png"
	} else {
		ext = ".jpg"
		// Re-read the image from the new buffer, and convert to jpg
		img := bimg.NewImage(buffer)
		buffer, err = img.Convert(bimg.JPEG)
	}

	return ext, buffer, err
}

func CompressImage(buffer []byte) ([]byte, error) {
	imgMeta, err := bimg.NewImage(buffer).Metadata()
	if err != nil {
		return nil, err
	}

	// Calculate new height and width
	height := imgMeta.Size.Height
	width := imgMeta.Size.Width
	size := bimg.ImageSize{}
	if height > *maxImgLength || width > *maxImgLength {
		size = GetNewImageSize(width, height)
	} else {
		size = imgMeta.Size
	}

	// Set options
	options := bimg.Options{}
	options.StripMetadata = true
	options.Quality = 100
	options.Compression = 3
	options.Height = size.Height
	options.Width = size.Width

	// Process options such as StripMetadata
	buffer, err = bimg.Resize(buffer, options)
	if err != nil {
		return nil, err
	}

	// Process image with new options
	img := bimg.NewImage(buffer)
	buffer, err = img.Process(options)

	return buffer, err
}

// TODO
//func RemoveExif(buffer []byte) {
//	img := bimg.NewImage(buffer)
//	imgMeta, err := img.Metadata()
//
//	imgMeta
//}

// GetNewImageSize will calculate a new ImageSize with a ratio as similar as possible to the original
// with the longest side set to *maxImgLength
func GetNewImageSize(width int, height int) bimg.ImageSize {
	heightF := float64(height)
	widthF := float64(width)
	maxLengthF := float64(*maxImgLength)
	max := math.Max(heightF, widthF)

	// Calculate decimal percentage change, eg 0.1 = -10%
	change := (max - maxLengthF) / max

	newHeight := heightF - (heightF * change)
	newWidth := widthF - (widthF * change)
	return bimg.ImageSize{Width: toInt(newWidth), Height: toInt(newHeight)}
}

func toInt(f float64) int {
	return int(f + 0.5)
}
