#!/usr/bin/env bash
set -eo pipefail

# Build
DOCKER_TAG="$(openssl rand -hex 8)"
export DOCKER_TAG
docker build . -t "firecow/ratnation:$DOCKER_TAG"

# Deploy
docker stack deploy -c examples/docker-swarm/stack.yml ratnation
