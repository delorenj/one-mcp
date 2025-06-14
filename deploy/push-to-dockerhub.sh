#!/bin/bash
set -e

VERSION=$(cat ../VERSION)
DOCKER_REPO="buru2020/one-mcp"

echo "Setting up Docker Buildx for multi-architecture builds..."
# 创建并使用多架构构建器（如果不存在）
if ! docker buildx ls | grep -q multiarch; then
    docker buildx create --use --name multiarch --driver docker-container
fi
docker buildx use multiarch

echo "Building and pushing multi-architecture images..."
echo "Building version: $VERSION"

# 构建并推送多架构镜像（同时推送 latest 和版本标签）
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t $DOCKER_REPO:latest \
    -t $DOCKER_REPO:$VERSION \
    --push \
    ../

echo "Successfully pushed multi-architecture images to Docker Hub!"
echo "Latest: https://hub.docker.com/r/$DOCKER_REPO/tags"
echo "Supported platforms: linux/amd64, linux/arm64"
echo ""
echo "You can now pull the image with:"
echo "  docker pull $DOCKER_REPO:latest"
echo "  docker pull $DOCKER_REPO:$VERSION"
echo ""
echo "Verifying multi-architecture manifest:"
docker buildx imagetools inspect $DOCKER_REPO:latest 