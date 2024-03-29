# Ratnation 

[![QA](https://github.com/firecow/ratnation/actions/workflows/quality-assurance.yml/badge.svg)](https://github.com/firecow/ratnation/actions/workflows/quality-assurance.yml)

Service mesh based on [rathole](https://github.com/rapiz1/rathole) and [traefik](https://github.com/traefik/traefik)

Consists of three different applications to operate

### ratcouncil
A service discovery application, used by ratkings and ratlings

### ratking
Controlplane application starting rathole servers, must be reachable for all ratlings

### ratling
Dataplane application managing rathole clients and traefik proxies, can be completely isolated

### encryption
Since reverse tunnel and proxy encryption isn't implemented yet, it's highly recommended that network traffic encryption is handled via other mechanisms (e.g. [Nebula](https://github.com/slackhq/nebula) or VPN), unless you are absolutely sure your traffic will stay in-house


## Quickstart

Install `docker` and call `docker swarm init`

See [stack.yml](./examples/docker-swarm/stack.yml) for deployment configuration

```bash
./example_deploy.sh
```

## Development

### Requirements
nodejs `>=18.x.x`

### Code reloading script

```
node src/start-dev.mjs
```
