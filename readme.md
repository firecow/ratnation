# Ratnation 

[![QA](https://github.com/firecow/ratnation/actions/workflows/quality-assurance.yml/badge.svg)](https://github.com/firecow/ratnation/actions/workflows/quality-assurance.yml)

Service mesh based on [rathole](https://github.com/rapiz1/rathole) and [traefik](https://github.com/traefik/traefik)

Consists of three different applications to operate

### ratcouncil
A service discovery application, used by ratkings and ratlings, must be exposed to the internet.

### ratking
Controlplane application starting rathole servers, must be exposed to the internet

### ratling
Dataplane application managing rathole clients and traefik proxies based


## Quickstart

Install `docker` and call `docker swarm init`

See [stack.yml](./examples/docker-swarm/stack.yml) for deployment configuration

```bash
./deploy_stack.sh
```

## Development

Starts applications with code reloading capabilities via nodemon

```
./start_dev.sh
```
