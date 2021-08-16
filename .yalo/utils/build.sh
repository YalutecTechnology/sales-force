#!/bin/sh

# build the Docker image
if [ -z ${BRANCH+x} ];
then
    export IMAGE_NAME=yalochat/$REPO:$(echo $BITBUCKET_BRANCH|sed 's#/#-#g')-$BITBUCKET_COMMIT
else
    export IMAGE_NAME=yalochat/$REPO:$(echo $BRANCH|sed 's#/#-#g')-$BITBUCKET_COMMIT
fi

echo "🚀 Image tag: $IMAGE_NAME"

echo "🚀 Docker build..."
docker build -t $IMAGE_NAME -f app/build/Dockerfile .
# authenticate with the Docker Hub registry

echo "🚀 Docker login..."
docker login --username $DOCKER_HUB_USER --password $DOCKER_HUB_PASSWORD

echo "🚀 Pushing Docker image..."
# push the new Docker image to the Docker registry
docker push $IMAGE_NAME
