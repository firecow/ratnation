# Ratnation 

[![QA](https://github.com/firecow/ratnation/actions/workflows/quality-assurance.yml/badge.svg)](https://github.com/firecow/ratnation/actions/workflows/quality-assurance.yml)

Service mesh based on [rathole](https://github.com/rapiz1/rathole) and [traefik](https://github.com/traefik/traefik)

Consists of three different applications to operate

### ratcouncil
A service discovery application, used by ratkings and ratlings, must be exposed to all ratlings and ratkings

### ratking
Controlplane application starting rathole servers, must be exposed to all ratlings

### ratling
Dataplane application managing rathole clients and traefik proxies, can be completely isolated.


## Quickstart

Install `docker` and call `docker swarm init`

See [stack.yml](./examples/docker-swarm/stack.yml) for deployment configuration

```bash
./deploy_stack.sh
```

## Development

Starts applications with code reloading capabilities

```
node src/start-dev-mjs
```
