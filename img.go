package main

import (
	"bytes"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/h2non/bimg"
	"github.com/valyala/fasthttp"
	"image"
	_ "image/jpeg"
	_ "image/png" // used by prominent color
	"io/ioutil"
	"log"
	"math"
	"os"
)

// ImageData TODO: image attribution and author
type ImageData struct {
	ImageName string `json:"image_name"`
	ImageUrl  string `json:"image_url"`
	Color     string `json:"median_color"`
}

// MainImageColor calculates the median color of cropped image
func MainImageColor(image string) string {
	img, err := LoadImage(image)
	if nil != err {
		if *removeBroken {
			err = os.Remove(image)
			log.Printf("failed to remove broken image: %s\n", err)
		}
		log.Printf("failed loading image %s: %s\n", image, err)
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

// LoadImage returns an image object for path
func LoadImage(path string) (image.Image, error) {
	imgBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("failed loading %s: %s", path, err)
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewBuffer(imgBytes))
	return img, err
}

// GetAllImages TODO: Maybe we could also cache the creation of ImageData?
// GetAllImages returns a slice of all images in the format of ImageData
func GetAllImages() []ImageData {
	images := make([]ImageData, 0)
	for _, hash := range imageCache {
		color := GetColor(hash)
		imageData := ImageData{
			ImageName: hash,
			ImageUrl:  liveUrl + "/images/" + hash,
			Color:     color,
		}
		images = append(images, imageData)
	}

	return images
}

// GetRandomImage TODO: Maybe we could also cache the creation of ImageData? As it is used for each request
// GetRandomImage chooses a random image name from the image cache
func GetRandomImage() ImageData {
	// rand.Intn panics if n <= 0.
	if len(imageCache) <= 0 {
		return ImageData{
			ImageName: "null",
			ImageUrl:  liveUrl + "/images/" + "null",
			Color:     "000000",
		}
	}

	hash := imageCache[rGen.Intn(len(imageCache))]
	color := GetColor(hash)

	imageData := ImageData{
		ImageName: hash,
		ImageUrl:  liveUrl + "/images/" + hash,
		Color:     color,
	}
	return imageData
}

// ImageHandler is our own RequestHandler with a CacheDuration of 0
func ImageHandler(root string, stripSlashes int) fasthttp.RequestHandler {
	fs := &fasthttp.FS{
		Root:               root,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: true,
		AcceptByteRange:    true,
		CacheDuration:      0,
	}
	if stripSlashes > 0 {
		fs.PathRewrite = fasthttp.NewPathSlashesStripper(stripSlashes)
	}
	return fs.NewRequestHandler()
}

// SaveFinal will compress and convert the image from path inside the config/tmp/ folder, and save the final
// result inside the config/images/ folder
func SaveFinal(path string) (string, error) {
	buffer, err := ConvertAndCompress(path)
	compressedPath := path + "-min"
	removePath := path

	fiOriginal, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	err = bimg.Write(compressedPath, buffer)
	if err != nil {
		return "", err
	}

	fiNew, err := os.Stat(compressedPath)
	if err != nil {
		return "", err
	}

	// We want to do this check in case the original image was more efficiently compressed than ours
	// fiNew is path + ext (compressed), fiOriginal is path
	if fiNew.Size() > fiOriginal.Size() {
		buffer, err = os.ReadFile(path)
		if err != nil {
			return "", err
		}

		// We only want to keep the smaller version if it is the correct file format
		imgType := bimg.DetermineImageType(buffer)
		if imgType == bimg.PNG || imgType == bimg.JPEG {
			removePath = path + "-min"
			compressedPath = path
		}
	}

	// Move compressed file to www/images/<file hash>
	hash, err := GetFileHash(compressedPath)
	if err != nil {
		return "", err
	}
	err = os.Rename(compressedPath, "config/images/"+hash)
	if err != nil {
		return "", err
	}

	err = os.Remove(removePath)

	return hash, err
}

// ConvertAndCompress will convert the image to jpg if it's non-transparent, and compress
// if it meets the requirements for being compressed
func ConvertAndCompress(path string) ([]byte, error) {
	buffer, err := bimg.Read(path)
	if err != nil {
		return nil, err
	}

	imgType, err := GetNewImageType(buffer)
	if err != nil {
		return nil, err
	}

	buffer, err = CompressImage(buffer, imgType)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

// GetNewImageType will take path and convert the image to a png
func GetNewImageType(buffer []byte) (bimg.ImageType, error) {
	img, err := bimg.NewImage(buffer).Metadata()
	if err != nil {
		return bimg.UNKNOWN, err
	}

	if img.Alpha == true {
		return bimg.PNG, nil
	} else {
		return bimg.JPEG, nil
	}
}

func CompressImage(buffer []byte, imgType bimg.ImageType) ([]byte, error) {
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
	options.Type = imgType

	// Process options such as StripMetadata
	buffer, err = bimg.Resize(buffer, options)
	if err != nil {
		return nil, err
	}

	// Process image with new options
	img := bimg.NewImage(buffer)
	buffer, err = img.Process(options)

	// Make sure the changes are now saved
	buffer = bimg.NewImage(buffer).Image()
	return buffer, err
}

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
	return bimg.ImageSize{Width: ToInt(newWidth), Height: ToInt(newHeight)}
}
