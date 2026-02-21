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
Rathole tunnel traffic between kings and lings can be encrypted using the [Noise protocol](https://noiseprotocol.org/) (NK pattern). Generate a keypair with `rathole --genkey` and pass the keys to the king:

```bash
king --noise-private-key="<base64-private-key>" --noise-public-key="<base64-public-key>" ...
```

The king sends its public key to the council, which distributes it to lings via state. Lings automatically enable encryption when connecting to kings that have a noise public key. Kings without noise keys continue to work unencrypted.

For proxy traffic or additional defense-in-depth, network-level encryption via [Nebula](https://github.com/slackhq/nebula) or VPN is still recommended.


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
