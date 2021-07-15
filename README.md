# srp-go
[![time tracker](https://wakatime.com/badge/github/l1ving/srp-go.svg)](https://wakatime.com/badge/github/l1ving/srp-go)
[![Docker Pulls](https://img.shields.io/docker/pulls/l1ving/srp-go?logo=docker&logoColor=white)](https://hub.docker.com/r/l1ving/srp-go)
[![Docker Build](https://img.shields.io/github/workflow/status/l1ving/srp-go/docker-build?logo=docker&logoColor=white)](https://github.com/l1ving/srp-go/actions/workflows/docker-build.yml)
[![CodeFactor](https://img.shields.io/codefactor/grade/github/l1ving/srp-go?logo=codefactor&logoColor=white)](https://www.codefactor.io/repository/github/l1ving/srp-go)

Serve a random image on `/` with a dynamic background color, let users upload on `/upload` and browse a gallery on `/browse`.
Has (optional) image resizing and compression.

This project was written in the span of about two working days, so there is likely code improvements to be made.
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
./srp-go -addr=localhost:6060
```

## TODO:

- [x] Add Makefile
  - [x] Add Dockerfile
  - [x] Add `.dockerignore`
  - [x] Add Github Actions
  - [ ] Add codefactor
  - [ ] Add badges
- [x] Add usage instructions
- [x] Finish `/upload` page
- [ ] Finish `/browse` page
- [ ] Add authentication for uploading
- [ ] Cleanup error handling
- [ ] Issues with missing folder on startup
  - [ ] Issues with no files on startup
- [ ] Issues with missing tmp folder
- [ ] Issues with safety when deleting an image
- [ ] Fix image color selection
