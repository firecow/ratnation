# Ratnation 

Service mesh based on [rathole](https://github.com/rapiz1/rathole) reverse tunnel application, ssh tunneling and socat

Consists of three different applications to operate

### ratnation council
A service discovery application, used by ratkings and ratlings, must be exposed to the internet.

### ratnation king
Controlplane application starting rathole servers based on council state, must be exposed to the internet

### ratnation ling
Dataplane application managing ratholes and socat proxies based on council state and command line options


## Quickstart

Install `docker` and call `docker swarm init`

See [stack.yml](./stack.yml) for deployment details

```bash
DOCKER_SWARN_NODE_IP=$(hostname -I | cut -d' ' -f1)
export DOCKER_SWARN_NODE_IP
echo "DOCKER_SWARN_NODE_IP=$DOCKER_SWARN_NODE_IP"
docker stack deploy -c stack.yml ratnation
```
