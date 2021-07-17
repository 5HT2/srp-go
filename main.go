package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	addr         = flag.String("addr", "localhost:6060", "TCP address to listen to")
	useTls       = flag.Bool("tls", false, "Whether to enable TLS")
	tlsCert      = flag.String("cert", "", "Full certificate file path")
	tlsKey       = flag.String("key", "", "Full key file path")
	maxImgLength = flag.Int("maximglength", 2000, "Maximum image height and width")
	maxBodySize  = flag.Int("maxbodysize", 100*1024*1024, "MaxRequestBodySize, defaults to 100MiB")

	images                    = UpdateImageCache()
	currentImage, currentHash = GetRandomImage()
	currentImageColor         = "000000"
	currentCss                = ""
	cssPath                   = []byte("/css/")
	apiPath                   = []byte("/api/")
	imgPath                   = []byte("/image")
	faviconPath               = []byte("/favicon.ico")
	rootStylePath             = []byte("/css/style.css")

	imgHandler = ImageHandler("www/images")
	cssHandler = fasthttp.FSHandler("www/css", 1)

	rSrc = rand.NewSource(time.Now().Unix())
	rGen = rand.New(rSrc) // initialize local pseudorandom generator
)

func main() {
	flag.Parse()
	log.Print("- Loading srp-go")
	go StartPolling()
	MkImageFolder()
	InitHtml()

	protocol := "http"
	if *useTls {
		protocol += "s"
	}

	log.Printf("- Running srp-go on " + protocol + "://" + *addr)

	s := &fasthttp.Server{
		Handler:            requestHandler,
		Name:               "srp-go",
		MaxRequestBodySize: *maxBodySize,
	}

	if *useTls && len(*tlsCert) > 0 && len(*tlsKey) > 0 {
		if err := s.ListenAndServeTLS(*addr, *tlsCert, *tlsKey); err != nil {
			log.Fatalf("- Error in ListenAndServeTLS: %s", err)
		}
	} else {
		if err := s.ListenAndServe(*addr); err != nil {
			log.Fatalf("- Error in ListenAndServe: %s", err)
		}
	}
}

// Main request handler
func requestHandler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	SetCacheHeaders(ctx)
	fmt.Println("")
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		log.Printf("%v: %v", string(key), string(value))
	})

	switch {
	// Serve an image on /image
	case bytes.Equal(path, imgPath):
		ctx.Response.Header.Set("X-Image-Hash", currentHash)

		ctx.URI().SetPathBytes([]byte("/" + currentImage))
		imgHandler(ctx)

	// Server 404 on /favicon.ico TODO: Add a default favicon
	case bytes.Equal(path, faviconPath):
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNotFound)

	// Serve css on /css/style.css
	case bytes.Equal(path, rootStylePath):
		ctx.Response.Header.Set("Content-Type", "text/css; charset=utf-8")
		_, _ = fmt.Fprint(ctx, currentCss)

	// Handle alternate css styles on /css/
	case bytes.HasPrefix(path, cssPath):
		cssHandler(ctx)

	// Handle the api on /api/
	case bytes.HasPrefix(path, apiPath):
		HandleApi(ctx)

	// Default to serving html on all other paths
	default:
		content := GetHtml(ctx)
		if len(content) > 0 {
			_, _ = fmt.Fprint(ctx, content)
		}
	}
}

// SetCacheHeaders sets the headers to avoid caching
func SetCacheHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Pragma-directive", "no-cache")
	ctx.Response.Header.Set("Cache-directive", "no-cache")
	ctx.Response.Header.Set("Cache-control", "no-store") // firefox ignores no-cache for this one
	ctx.Response.Header.Set("Pragma", "no-cache")
	ctx.Response.Header.Set("Expires", "0")
	ctx.Response.Header.Set("ETag", strconv.FormatInt(time.Now().UnixNano(), 10))
}

// UpdateCurrentCss updates currentCss with 000000 replaced with currentImageColor
func UpdateCurrentCss() {
	content, err := ioutil.ReadFile("www/css/style.css")
	if err != nil {
		log.Fatal(err)
	}

	contentStr := strings.Replace(string(content), "000000", currentImageColor, 1)
	currentCss = contentStr
}
