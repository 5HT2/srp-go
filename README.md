# [srp-go](https://frog.pics)

[![time tracker](https://wakatime.com/badge/github/l1ving/srp-go.svg)](https://wakatime.com/badge/github/l1ving/srp-go)
[![Docker Pulls](https://img.shields.io/docker/pulls/l1ving/srp-go?logo=docker&logoColor=white)](https://hub.docker.com/r/l1ving/srp-go)
[![Docker Build](https://img.shields.io/github/workflow/status/l1ving/srp-go/docker-build?logo=docker&logoColor=white)](https://github.com/l1ving/srp-go/actions/workflows/docker-build.yml)
[![CodeFactor](https://img.shields.io/codefactor/grade/github/l1ving/srp-go?logo=codefactor&logoColor=white)](https://www.codefactor.io/repository/github/l1ving/srp-go)

Serve a random image on [\<main page\>](https://frog.pics) with a dynamic background color, let users upload
on [/upload](https://frog.pics/upload) and browse a gallery on [/browse](https://frog.pics/browse). Has (optional) image
resizing and compression.

Contributions are welcome and appreciated.

## Contributing

To build:

```bash
git clone git@github.com:l1ving/srp-go.git
cd srp-go
make
```

To run:

```bash
./srp-go -h # for a full list of parameters (none required)
./srp-go -addr=localhost:6060 -debug=true
```

I recommend deleting [`sample.jpg`](https://github.com/l1ving/srp-go/blob/master/config/images/sample.jpg)
after you have uploaded a few pictures. The sample file is there to prevent issues when first testing
(needing an image to serve on `/`, to generate a sample color for, etc).

## OAuth

In order to set up OAuth, follow [these](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app)
instructions to create a GitHub OAuth app.

You will want to 
- Set `LIVE_URL` to the accessible url of your site
  - Eg: `https://frog.pics`
  - Or: `http://localhost:6060` for a testing environment
- Enable "Request user authorization (OAuth)"
- Enable the Read-Only option for User Email addresses (aka `user:email`)

Create a `.env` file inside your config folder, with the following format:
```bash
OAUTH_CLIENT_ID=Iv1.some_client_id
OAUTH_CLIENT_SECRET=your_client_secret
LIVE_URL=http://localhost:6060
WEBHOOK_URL= # optional discord webhook url for posting specific events
```

## API

The full list of accessible API endpoints can be found inside [`api.go`](https://github.com/l1ving/srp-go/blob/master/api.go).

The `/api/random` endpoint will return the properties of a randomly-selected image in json (by default), like so:
```json
{
    "image_name": "sample.jpg",
    "image_url": "http://localhost:6060/images/sample.jpg",
    "median_color": "868232"
}
```

You can also add `?format=css` to get the css version if you really want:
```css
body {
    background-color: #868232;
}

div.img {
    content: url('/images/sample.jpg');
}
```

## TODO:

- [ ] Add authentication for uploading
- [ ] Add Discord embeds
- [ ] Switch to a different prominent color library, because the current one doesn't support webp
  - [ ] Switch to webp for saving images
- [ ] Finalize webhook support
- [ ] Image attribution support

## License

This project is licensed under [ISC](https://github.com/l1ving/srp-go/blob/master/LICENSE.md).

The [`sample.jpg`](https://github.com/l1ving/srp-go/blob/master/config/images/sample.jpg) file is licensed under
Creative Commons Attribution-Share Alike, you can find the original file
[here](https://commons.wikimedia.org/wiki/File:Bufo_americanus_PJC1.jpg).
