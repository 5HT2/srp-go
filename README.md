# [srp-go](https://frog.pics)
[![time tracker](https://wakatime.com/badge/github/l1ving/srp-go.svg)](https://wakatime.com/badge/github/l1ving/srp-go)
[![Docker Pulls](https://img.shields.io/docker/pulls/l1ving/srp-go?logo=docker&logoColor=white)](https://hub.docker.com/r/l1ving/srp-go)
[![Docker Build](https://img.shields.io/github/workflow/status/l1ving/srp-go/docker-build?logo=docker&logoColor=white)](https://github.com/l1ving/srp-go/actions/workflows/docker-build.yml)
[![CodeFactor](https://img.shields.io/codefactor/grade/github/l1ving/srp-go?logo=codefactor&logoColor=white)](https://www.codefactor.io/repository/github/l1ving/srp-go)

Serve a random image on [\<main page\>](https://frog.pics) with a dynamic background color, let users upload on [/upload](https://frog.pics/upload) and browse a gallery on [/browse](https://frog.pics/browse).
Has (optional) image resizing and compression.

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

## TODO:

- [x] Add Makefile
  - [x] Add Dockerfile
  - [x] Add `.dockerignore`
  - [x] Add Github Actions
  - [x] Add codefactor
  - [x] Add badges
- [x] Add usage instructions
- [x] Finish `/upload` page
- [x] Finish `/browse` page
- [ ] Add authentication for uploading
- [ ] Cleanup error handling
- [ ] Issues with missing folder on startup
  - [ ] Issues with no files on startup
- [x] Issues with missing tmp folder
- [ ] Issues with missing img folder
- [x] Issues with safety when deleting an image
- [x] Fix image color selection
- [ ] Add Discord embeds
- [ ] Add favicon
- [ ] Switch to a different prominent color library, because the current one doesn't support webp
  - [ ] Switch to webp for saving images
