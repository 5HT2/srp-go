package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	addr           = flag.String("addr", "localhost:6060", "TCP address to listen to")
	useTls         = flag.Bool("tls", false, "Whether to enable TLS")
	tlsCert        = flag.String("cert", "", "Full certificate file path")
	tlsKey         = flag.String("key", "", "Full key file path")
	maxImgLength   = flag.Int("maximglength", 2000, "Maximum image height and width")
	maxBodySize    = flag.Int("maxbodysize", 100*1024*1024, "MaxRequestBodySize, defaults to 100MiB")
	browseImgColor = flag.String("browseimgcolor", "eaffe8", "Background color to use for browse page")
	debug          = flag.Bool("debug", false, "Enable debug logging")
	liveUrl        = ""
	// TODO: Remove when auth added
	allowUpload = flag.Bool("allowupload", false, "Allow disabling upload (temporary)")

	cssPath     = []byte("/css/")
	svgPath     = []byte("/svg/")
	apiPath     = []byte("/api/")
	imgPath     = []byte("/images/")
	faviconPath = []byte("/favicon.ico")

	imgHandler = ImageHandler("config/images/", 1)

	rSrc = rand.NewSource(time.Now().Unix())
	rGen = rand.New(rSrc) // initialize local pseudorandom generator
)

func main() {
	flag.Parse()
	log.Print("- Loading srp-go")
	setup()
	log.Printf("- Running srp-go on %s", liveUrl)

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
	setCacheHeaders(ctx)
	handleDebug(ctx)

	switch {
	// Serve favicon on /favicon.ico
	case bytes.Equal(path, faviconPath):
		HandleCachedFavicon(ctx)

	// Handle css styles on /css/
	case bytes.HasPrefix(path, cssPath):
		content := GetCachedContent(ctx, cssMime)
		if len(content) > 0 {
			_, _ = fmt.Fprint(ctx, content)
		}

	// Handle svg files on /svg/
	case bytes.HasPrefix(path, svgPath):
		content := GetCachedContent(ctx, svgMime)
		if len(content) > 0 {
			_, _ = fmt.Fprint(ctx, content)
		}

	// Serve images on /images/
	case bytes.HasPrefix(path, imgPath):
		imgHandler(ctx)

	// Handle the api on /api/
	case bytes.HasPrefix(path, apiPath):
		HandleApi(ctx)

	// Default to serving html on all the other paths
	default:
		content := GetCachedContent(ctx, htmlMime)
		if len(content) > 0 {
			_, _ = fmt.Fprint(ctx, content)
		}
	}
}

// setCacheHeaders sets the headers to avoid caching
func setCacheHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Pragma-Directive", "no-cache")
	ctx.Response.Header.Set("Cache-Directive", "no-cache")
	ctx.Response.Header.Set("Cache-Control", "no-store") // firefox ignores no-cache for this one
	ctx.Response.Header.Set("Pragma", "no-cache")
	ctx.Response.Header.Set("Expires", "0")
	ctx.Response.Header.Set("ETag", strconv.FormatInt(time.Now().UnixNano(), 10))
}

// handleDebug will print the debugging information on requests, including ctx.Path() and headers
func handleDebug(ctx *fasthttp.RequestCtx) {
	if *debug {
		fileCache = LoadAllCaches()       // re-read html caches for easier debugging, worse performance
		galleryCache = LoadGalleryCache() // re-create gallery html
		fmt.Println("")
		log.Printf("Path: %s", ctx.Path())
		ctx.Request.Header.VisitAll(func(key, value []byte) {
			log.Printf("%v: %v", string(key), string(value))
		})
	}
}

// setup will run whatever setup functions should be performed upon starting srp-go
func setup() {
	// Load env variables
	err := godotenv.Load("config/.env")
	checkMissingDirs()
	// Fix missing dirs
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
	// Set config options now that the env is loaded
	liveUrl = os.Getenv("LIVE_URL")
	webhookUrl = os.Getenv("WEBHOOK_URL")

	// Set the proper oauthConfig now that flags and env have been loaded
	oauthClient = os.Getenv("OAUTH_CLIENT_ID")
	oauthSecret = os.Getenv("OAUTH_CLIENT_SECRET")
	oauthConfig = &oauth2.Config{
		RedirectURL:  liveUrl + "/api/auth/callback",
		ClientID:     oauthClient,
		ClientSecret: oauthSecret,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

// checkMissingDirs will check for missing directories (side effect from git, empty dirs don't get committed)
func checkMissingDirs() {
	// Check if tmp folder exists. Technically only needed for non-Docker testing
	if _, err := os.Stat("config/tmp/"); os.IsNotExist(err) {
		if err != nil {
			log.Printf("- Error checking for config/tmp/ folder: %v", err)
		}

		err = os.Mkdir("config/tmp/", os.FileMode(0700))
		if err != nil {
			log.Fatalf("- Error making config/tmp/ folder: %v", err)
			return
		}

		log.Printf("- Created config/tmp/ folder")
	}
}
