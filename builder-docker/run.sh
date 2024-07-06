#!/usr/bin/env sh

cd $(dirname $0)

clear
echo "Checking if the Docker image exists..."

if docker images | grep -q mkvtool-builder; then
  echo "Docker image mkvtool-builder exists. Running container..."
else
  echo "Docker image mkvtool-builder does not exist. Building image..."
  docker build -q -t mkvtool-builder . > /dev/null 2>&1
  echo "Docker image built successfully."
fi

echo "Running container..."
docker run --rm -it -v ./dist:/dist mkvtool-builder
echo "Container finished. Exiting..."
