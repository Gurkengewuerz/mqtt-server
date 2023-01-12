#!/bin/bash

REGISTRY="docker.io"
USERNAME="gurken2108"
PROJECT="mqtt-server"

#docker buildx create --use --name ${USERNAME}-${PROJECT}
#docker buildx build \
#  --platform linux/amd64,linux/arm/v7,linux/arm64 \
#  --push \
#  -t ${REGISTRY}/${USERNAME}/${PROJECT}:latest \
#  .
#docker buildx stop ${USERNAME}-${PROJECT}
#docker buildx rm ${USERNAME}-${PROJECT}

docker build --no-cache -t ${REGISTRY}/${USERNAME}/${PROJECT}:latest .
#docker push ${REGISTRY}/${USERNAME}/${PROJECT}:latest
