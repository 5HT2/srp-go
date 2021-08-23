package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
	"strings"
)

func HandleGeneric(ctx *fasthttp.RequestCtx, status int, message string) {
	ctx.Response.SetStatusCode(status)
	ctx.Response.Header.Set("X-Server-Message", message)
	ctx.Response.Header.Set("Content-Type", "text/plain")
	fmt.Fprintf(ctx, "%v %s\n", status, message)
}

func HandleNotFound(ctx *fasthttp.RequestCtx) {
	HandleGeneric(ctx, fasthttp.StatusNotFound, "Not Found")
}

func HandleWrongMethod(ctx *fasthttp.RequestCtx) {
	HandleGeneric(ctx, fasthttp.StatusMethodNotAllowed, "Cannot "+string(ctx.Method())+" on "+string(ctx.Path()))
}

func HandleForbidden(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
	ctx.Response.Header.Set("X-Server-Message", "403 Forbidden")
	ctx.Response.Header.Set("Content-Type", "text/plain")
	fmt.Fprint(ctx, "403 Forbidden\n")
	log.Printf("- Returned 403 to %s - tried to connect with '%s' to '%s'",
		ctx.RemoteIP(), ctx.Request.Header.Peek("Auth"), ctx.Path())
}

func HandleInternalServerError(ctx *fasthttp.RequestCtx, message string, err error) {
	if strings.HasSuffix(err.Error(), "no such file or directory") {
		HandleNotFound(ctx)
		return
	}

	ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.Response.Header.Set("X-Server-Message", "500 "+err.Error())
	ctx.Response.Header.Set("Content-Type", "text/plain")
	fmt.Fprintf(ctx, "500 %v\n", err)
	log.Printf("- %v: %s %v", ctx.RemoteIP(), message, err)
}
