#!/bin/sh

# get the image that was built in the master branch
export IMAGE_NAME=yalochat/$REPO:staging-$BITBUCKET_TAG
export NEW_IMAGE_NAME=yalochat/$REPO:production-$BITBUCKET_TAG

echo "🚀 Image name: $IMAGE_NAME"
echo "🚀 New image name: $NEW_IMAGE_NAME"

echo "🚀 Docker login..."
# authenticate with the Docker Hub registry
docker login --username $DOCKER_HUB_USER --password $DOCKER_HUB_PASSWORD

echo "🚀 Docker pull $IMAGE_NAME"
# pull the image down
docker pull $IMAGE_NAME

echo "🚀 Docker tag $IMAGE_NAME $NEW_IMAGE_NAME"
# retag the image using the git tag
docker tag $IMAGE_NAME $NEW_IMAGE_NAME

echo "🚀 Docker push $NEW_IMAGE_NAME..."
# push the image back
docker push $NEW_IMAGE_NAME
