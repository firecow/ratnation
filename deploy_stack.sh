DOCKER_SWARN_NODE_IP=$(hostname -I | cut -d' ' -f1)
export DOCKER_SWARN_NODE_IP
echo "DOCKER_SWARN_NODE_IP=$DOCKER_SWARN_NODE_IP"
docker stack deploy -c stack.yml ratnation
