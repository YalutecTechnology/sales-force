#!/bin/sh

docker_image_tag_exists() {
    curl --silent -u $DOCKER_HUB_USER:$DOCKER_HUB_PASSWORD -f -lSL https://index.docker.io/v1/repositories/$1/tags/$2 > /dev/null
}

# Whenever the tag is first pushed, the Docker image is built
# and tagged with staging-$BITBUCKT_TAG. Weve also marked this step with
# deployment: staging so that it will show up in the deployments window
#
# build the Docker image
export NEW_IMAGE_NAME=yalochat/$REPO:staging-$BITBUCKET_TAG

if docker_image_tag_exists yalochat/$REPO master-$BITBUCKET_COMMIT; then
    export IMAGE_NAME=yalochat/$REPO:master-$BITBUCKET_COMMIT
else
    export IMAGE_NAME=yalochat/$REPO:release-$BITBUCKET_COMMIT
fi

echo "🚀 Image name: $IMAGE_NAME"
echo "🚀 New image name: $NEW_IMAGE_NAME"

echo "🚀 Docker login..."
docker login --username $DOCKER_HUB_USER --password $DOCKER_HUB_PASSWORD

echo "🚀 Docker pull $IMAGE_NAME"
docker pull $IMAGE_NAME

echo "🚀 Docker tag $IMAGE_NAME $NEW_IMAGE_NAME"
docker tag $IMAGE_NAME $NEW_IMAGE_NAME

echo "🚀 Docker push $NEW_IMAGE_NAME..."
# push the new Docker image to the Docker registry
docker push $NEW_IMAGE_NAME
