NAME   := l1ving/srp-go
TAG    := $(shell git log -1 --pretty=%h)
IMG    := ${NAME}:${TAG}
LATEST := ${NAME}:latest

srp-go: clean
	go get -u github.com/valyala/fasthttp
	go get -u github.com/h2non/bimg
	go get -u github.com/EdlinOrg/prominentcolor
	go get -u golang.org/x/oauth2
	go build

clean:
	rm -f srp-go

build:
	@docker build -t ${IMG} .
	@docker tag ${IMG} ${LATEST}

push:
	@docker push ${NAME}

login:
	@docker login -u ${DOCKER_USER} -p ${DOCKER_PASS}
