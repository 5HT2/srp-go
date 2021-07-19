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

I recommend deleting [`sample.jpg`](https://github.com/l1ving/srp-go/blob/master/www/content/images/sample.jpg)
after you have uploaded a few pictures. The sample file is there to prevent issues when first testing
(needing an image to serve on `/`, to generate a sample color for, etc).

## TODO:

- [ ] Add authentication for uploading
- [ ] Add Discord embeds
- [ ] Add favicon
- [ ] Switch to a different prominent color library, because the current one doesn't support webp
    - [ ] Switch to webp for saving images

## License

This project is licensed under [ISC](https://github.com/l1ving/srp-go/blob/master/LICENSE.md).

The [`sample.jpg`](https://github.com/l1ving/srp-go/blob/master/www/content/images/sample.jpg) file is licensed under
Creative Commons Attribution-Share Alike, you can find the original file
[here](https://commons.wikimedia.org/wiki/File:Bufo_americanus_PJC1.jpg).
