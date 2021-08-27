package main

import (
	"bytes"
	"encoding/base64"
	json2 "encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"golang.org/x/oauth2"
	"log"
	"math/rand"
	"strconv"
	"time"
)

var (
	authPath         = "/api/auth"
	authCallbackPath = "/api/auth/callback"
	authVerifyPath   = "/api/auth/verify"
	randomPath       = "/api/random"
	uploadPath       = "/api/upload"

	oauthConfig = &oauth2.Config{} // Set in main.go after flags have been parsed
	oauthClient = ""               // Set in main.go after env is parsed
	oauthSecret = ""               // Set in main.go after env is parsed
	cookieName  = "OAuth-State"

	ghAccessTokenUrl = "https://github.com/login/oauth/access_token"
	ghApiUserUrl     = "https://api.github.com/user"
)

type ghAuthResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
}

type ghUserResponse struct {
	Id        int    `json:"id"`
	Username  string `json:"login"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
	HtmlUrl   string `json:"html_url"`
}

func HandleApi(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())

	switch path {
	case authPath, authCallbackPath, randomPath:
		if !ctx.IsGet() {
			HandleWrongMethod(ctx)
			return
		}
	case uploadPath, authVerifyPath:
		if !ctx.IsPost() {
			HandleWrongMethod(ctx)
			return
		}
	}

	switch path {
	case uploadPath:
		handleUpload(ctx)
	case authPath:
		handleAuth(ctx)
	case authCallbackPath:
		handleAuthCallback(ctx)
	case authVerifyPath:
		handleAuthVerify(ctx)
	case randomPath:
		handleRandom(ctx)
	default:
		HandleNotFound(ctx)
	}
}

// handleRandom handles the query parameters for /api/random
func handleRandom(ctx *fasthttp.RequestCtx) {
	format := string(ctx.FormValue("format"))

	if format == "css" {
		handleDynamicCss(ctx)
	} else {
		handleRandomJson(ctx)
	}
}

// handleDynamicCss returns the css to insert a random image with its background color onto the main page
func handleDynamicCss(ctx *fasthttp.RequestCtx) {
	image, color := GetRandomImage()

	// this is probably the "easiest" way to do it without modifying html... use a dynamic @import stylesheet
	ctx.Response.Header.SetContentType(jsonMime)
	_, _ = fmt.Fprintf(ctx,
		"body {\n    background-color: #%s;\n}\n\ndiv.img {\n    content: url('/images/%s');\n}\n",
		color, image)
}

// handleRandomJson returns the json form of a random image, usually displayed on the main page
func handleRandomJson(ctx *fasthttp.RequestCtx) {
	image, color := GetRandomImage()

	// TODO: image attribution and author
	body := map[string]string{
		"image_name":   image,
		"image_url":    liveUrl + "/images/" + image,
		"median_color": color}
	json, err := json2.MarshalIndent(body, "", "    ")

	if err != nil {
		HandleInternalServerError(ctx, "Error formatting json", err)
		return
	}

	ctx.Response.Header.SetContentType(jsonMime)
	_, _ = fmt.Fprintf(ctx, "%s\n", json)
}

// handleAuth creates a cookie, sets the state and redirects the user to their auth code
func handleAuth(ctx *fasthttp.RequestCtx) {
	cookie, state := generateAuthCookie()
	url := oauthConfig.AuthCodeURL(state)
	ctx.Response.Header.SetCookie(cookie)
	ctx.Redirect(url, fasthttp.StatusTemporaryRedirect)
}

// handleAuthCallback handles the redirect after a successful auth from GitHub, and creatures a User in the database
func handleAuthCallback(ctx *fasthttp.RequestCtx) {
	code := ctx.FormValue("code")
	state := ctx.FormValue("state")

	// Make sure both values are set
	if len(code) == 0 || len(state) == 0 {
		HandleGeneric(ctx, fasthttp.StatusBadRequest, "Empty key 'code' or 'state'")
		return
	}

	cookieBytes := ctx.Request.Header.Cookie(cookieName)

	// Make sure their cookie state is the same one as returned by GitHub
	if !bytes.Equal(cookieBytes, state) {
		HandleForbidden(ctx)
		return
	}

	ghAuthRes, err := generateGithubAuthResponse(ctx, string(code))
	if err != nil {
		return // err is handled inside method
	}

	// GitHub returns a response code of 200 even when there's an error, so we have to check the string itself
	if ghAuthRes.Error != "" {
		HandleGeneric(ctx, fasthttp.StatusBadRequest, ghAuthRes.Error+": "+ghAuthRes.ErrorDescription)
		return
	}

	body, err := getGithubData(ctx, ghAuthRes.AccessToken)
	if err != nil {
		return // err is handled inside method
	}

	var ghUser ghUserResponse
	err = json2.Unmarshal(body, &ghUser)
	if err != nil {
		HandleInternalServerError(ctx, "Error decoding user response json", err)
		return
	}

	// We have successfully authenticated with GitHub now... redirect back to the /upload page
	ctx.Redirect("/upload", fasthttp.StatusTemporaryRedirect)
	PostMessage(ctx, ghUser)

	user := User{
		ghUser.Id,
		ghUser.Username,
		ghUser.Name,
		string(state),
		false,
	}
	err = InsertUser(user)
	if err != nil {
		log.Printf("- Failed to create new user: %v", err)
		// TODO: Handle this in the bot
	} else {
		saveDatabase()
	}
	// TODO: implement webhook posting and "logged in page", this only prints the users information currently
}

// handleAuthVerify handles a request which is supposed to contain a Cookie header with an OAuth-State in it
// and validates said cookie. Returns the user and whether it handled the request.
func handleAuthVerify(ctx *fasthttp.RequestCtx) (*User, bool) {
	stateBytes := ctx.Request.Header.Cookie(cookieName)
	if len(stateBytes) == 0 { // No cookie header
		HandleGeneric(ctx, fasthttp.StatusUnauthorized, "Not logged in!")
		return nil, true
	}

	user := GetUser(string(stateBytes))
	if user == nil { // A user with the state returned by the cookie was not found in the database
		HandleForbidden(ctx)
		return nil, true
	}

	// else, we don't do anything if the user is found. The request will return 200
	return user, false
}

func getGithubData(ctx *fasthttp.RequestCtx, accessToken string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	req.Header.Set("Accept", jsonMime)
	req.Header.Set("Authorization", "token "+accessToken)
	req.SetRequestURI(ghApiUserUrl)
	res := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, res); err != nil {
		fasthttp.ReleaseRequest(req)
		HandleInternalServerError(ctx, "Error getting Github user from Github API", err)
		return *&[]byte{}, err
	}
	fasthttp.ReleaseRequest(req)
	resBody := res.Body()
	body := make([]byte, len(resBody))
	copy(body, resBody)           // copy bytes to unused var
	fasthttp.ReleaseResponse(res) // When done with resBody

	return body, nil
}

func generateGithubAuthResponse(ctx *fasthttp.RequestCtx, code string) (ghAuthResponse, error) {
	body := map[string]string{"client_id": oauthClient, "client_secret": oauthSecret, "code": code}
	json, _ := json2.Marshal(body)

	req := fasthttp.AcquireRequest()
	req.SetBody(json)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType(jsonMime)
	req.Header.Set("Accept", jsonMime)
	req.SetRequestURI(ghAccessTokenUrl)
	res := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, res); err != nil {
		fasthttp.ReleaseRequest(req)
		HandleInternalServerError(ctx, "Error generating access token", err)
		return *&ghAuthResponse{}, err
	}
	fasthttp.ReleaseRequest(req)
	resBody := res.Body()

	var ghRes ghAuthResponse
	err := json2.Unmarshal(resBody, &ghRes)
	if err != nil {
		HandleInternalServerError(ctx, "Error decoding response json", err)
		return *&ghAuthResponse{}, err
	}

	fasthttp.ReleaseResponse(res) // When done with resBody
	return ghRes, nil
}

// generateAuthCookie will create an OAuth-State cookie with a 1-year expiry
func generateAuthCookie() (*fasthttp.Cookie, string) {
	expiration := time.Now().Add(365 * 24 * time.Hour)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := fasthttp.Cookie{}
	//cookie.SetSecure(true) // TODO uncomment in prod
	cookie.SetKey(cookieName)
	cookie.SetValue(state)
	cookie.SetExpire(expiration)
	cookie.SetHTTPOnly(true)
	cookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	cookie.SetPathBytes(apiPath) // We want the cookie to be accessible by /api/
	return &cookie, state
}

// handleUpload handles uploading one file per request
func handleUpload(ctx *fasthttp.RequestCtx) {
	user, alreadyHandled := handleAuthVerify(ctx)
	if alreadyHandled {
		return
	} else if user.Whitelisted != true {
		HandleGeneric(ctx, fasthttp.StatusForbidden, "Not whitelisted!")
		return
	}

	fh, err := ctx.FormFile("file")
	tmpName := strconv.FormatInt(time.Now().UnixNano(), 10)
	path := "config/tmp/" + tmpName

	if err == nil {
		err = fasthttp.SaveMultipartFile(fh, path)
		if err != nil {
			HandleInternalServerError(ctx, "Error saving file from /api/upload", err)
			return
		}

		image, err := SaveFinal(path)
		if err != nil {
			HandleInternalServerError(ctx, "Error converting file from /api/upload", err)
			return
		}

		ctx.Response.Header.Set("X-Image-Hash", image)
		HandleGeneric(ctx, fasthttp.StatusCreated, "Created")

		// Update image cache after uploading a new image
		// we want to check if it's missing in case the user uploads the same image more than once
		imageCache = AppendIfMissing(imageCache, image)
		// Update the browse gallery cache after uploading
		galleryCache = LoadGalleryCache()
	} else {
		HandleInternalServerError(ctx, "Other error with handling upload", err)
	}
}
