#!/bin/bash

# shellcheck disable=SC1091
source "$HOME/.env"
if [[ -z "$SRP_PATH" ]]; then
  echo "SRP_PATH not set!"
  exit 1
fi

docker pull l1ving/srp-go:latest

if [[ "$1" != "FIRST_RUN" ]]; then
  docker stop srp-go || echo "Could not stop missing container srp-go"
  docker rm srp-go || echo "Could not remove missing container srp-go"
fi

docker run --name srp-go \
  -e ADDRESS="localhost:6012" \
  --mount type=bind,source="$SRP_PATH",target=/srp-go/config \
  --network host -d \
  l1ving/srp-go
