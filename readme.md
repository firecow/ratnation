# Ratnation 
!! WIP !!


Service mesh based on the excellent [rathole](https://github.com/rapiz1/rathole) reverse tunnel application

Consists of three different applications to operate

### [ratcouncil](https://github.com/firecow/ratcouncil) 
A service discovery application, used by ratkings and ratunderlings, must be exposed to the internet.

### [ratking](https://github.com/firecow/ratking)
Controlplane application starting rathole servers based on council state, must be exposed to the internet

### [ratunderling](https://github.com/firecow/ratunderling) 
Dataplane application starting rathole clients and socat proxies based on council state and command line options


## Quickstart

Install `docker` and initialize a swarm `docker swarm init`

See [stack.yml](./stack.yml) for deployment details

```bash
export HOSTNAME=$(docker info --format '{{ .Swarm.NodeAddr }}')
docker stack deploy -c stack.yml ratnation
```
