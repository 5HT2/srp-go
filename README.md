# srp-go

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
