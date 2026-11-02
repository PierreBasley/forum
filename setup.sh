#!/bin/sh

if [ -z "$1" ]; then
    echo "You can optionally specify the name of the image you want to create in the first argument to the script"
    IMAGE_NAME=forum:latest
else
    IMAGE_NAME="$1"
fi
echo "Image will be named \"$IMAGE_NAME\""

sudo docker build -f Dockerfile -t "$IMAGE_NAME" .
RESULT=$?
if [ "$RESULT" -ne 0 ]; then
    echo "Build failed with status $RESULT, exiting early"
    exit
fi

if [ -z "$2" ]; then
    echo "You can optionally specify the name of the container you want to run (2nd argument)"
    CONTAINER=forum_container
else
    CONTAINER="$2"
fi
echo "Running container as \"$CONTAINER\""

sudo docker-compose up -d

sudo docker ps -a
