#!/bin/bash

docker buildx build \
  -f Dockerfile \
  -t icemarkom/pitemp \
  --platform=linux/amd64,linux/arm64,linux/arm \
  --push \
  .
