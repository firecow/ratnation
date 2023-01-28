#!/usr/bin/env bash
set -eo pipefail

# Build
DOCKER_TAG="$(openssl rand -hex 8)"
export DOCKER_TAG
docker build . -t "firecow/ratnation:$DOCKER_TAG"

# Deploy
DOCKER_SWARN_NODE_IP=$(hostname -I | cut -d' ' -f1)
export DOCKER_SWARN_NODE_IP
echo "DOCKER_SWARN_NODE_IP=$DOCKER_SWARN_NODE_IP"
docker stack deploy -c examples/docker-swarm/stack.yml ratnation
