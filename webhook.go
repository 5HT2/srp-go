package main

import (
	json2 "encoding/json"
	"github.com/valyala/fasthttp"
	"log"
)

var (
	webhookUrl        = "" // Set in main.go after env is parsed
	embedSuccessColor = 4388248
	embedFailColor    = 16073282
)

func PostMessage(ctx *fasthttp.RequestCtx, user ghUserResponse) {
	if len(webhookUrl) == 0 {
		return // hasn't been set in config/.env
	}

	type discordAuthor struct {
		Name    string `json:"name"`
		Url     string `json:"url"`
		IconUrl string `json:"icon_url"`
	}
	type discordThumbnail struct {
		Url string `json:"url"`
	}
	type discordField struct {
		Title       string `json:"name"`
		Description string `json:"value"`
	}
	type discordEmbed struct {
		Title     string           `json:"title"`
		Color     int              `json:"color"`
		Author    discordAuthor    `json:"author"`
		Thumbnail discordThumbnail `json:"thumbnail"`
		Fields    []discordField   `json:"fields"`
	}
	type discordMessage struct {
		Content   string         `json:"content"`
		AvatarUrl string         `json:"avatar_url"`
		Username  string         `json:"username"`
		Embeds    []discordEmbed `json:"embeds"`
	}

	host := string(ctx.Host())
	avatar := liveUrl + "/favicon.ico"
	author := discordAuthor{host, liveUrl, avatar}
	thumbnail := discordThumbnail{user.AvatarUrl}
	fields := []discordField{{user.Name, user.HtmlUrl}}
	message := discordMessage{AvatarUrl: avatar, Username: host, Embeds: []discordEmbed{{
		"User successfully authenticated with " + host,
		embedSuccessColor,
		author,
		thumbnail,
		fields,
	}}}

	json, err := json2.Marshal(message)

	if err != nil {
		if *debug {
			log.Printf("Error creating webhook message json: %s", err)
		}
		return // If this method fails, it doesn't matter to users
	}

	if *debug {
		log.Printf("Posting webhook json: %s", json)
	}

	req := fasthttp.AcquireRequest()
	req.SetBody(json)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType(jsonMime)
	req.SetRequestURI(webhookUrl)
	res := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, res); err != nil {
		fasthttp.ReleaseRequest(req)
		log.Printf("Error posting webhook: %s", err)
		return
	}
	fasthttp.ReleaseRequest(req)
	resBody := res.Body()
	if *debug && len(resBody) > 0 {
		log.Printf("Webhook response: %s", resBody)
	}
	fasthttp.ReleaseResponse(res) // When done with resBody
}
