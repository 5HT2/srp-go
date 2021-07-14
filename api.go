package main

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

var (
	uploadPath = []byte("/api/upload")
)

func HandleApi(ctx *fasthttp.RequestCtx) {
	if !ctx.IsPost() {
		HandleGeneric(ctx, fasthttp.StatusMethodNotAllowed, "Cannot "+string(ctx.Method())+" on /api/")
		return
	}

	path := ctx.Path()

	switch {
	case bytes.Equal(path, uploadPath):
		handleUpload(ctx)
	}
}

func handleUpload(ctx *fasthttp.RequestCtx) {
	fh, err := ctx.FormFile("file")
	tmpName := strconv.FormatInt(time.Now().UnixNano(), 10)
	path := "www/tmp/" + tmpName

	if err == nil {
		err = fasthttp.SaveMultipartFile(fh, path)
		if err != nil {
			HandleInternalServerError(ctx, err)
			return
		}

		err = SaveFinal(path)
		if err != nil {
			HandleInternalServerError(ctx, err)
			return
		}

		UpdateImageCache()
		HandleGeneric(ctx, fasthttp.StatusCreated, "Created")
	} else {
		HandleInternalServerError(ctx, err)
	}
}
