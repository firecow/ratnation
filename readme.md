# Ratnation 
!! WIP !!


A service mesh based on the excellent [rathole](https://github.com/rapiz1/rathole) reverse tunnel application

Consists of three different applications to operate

## [ratcouncil](https://github.com/firecow/ratcouncil) 
A service discovery application used by ratkings and ratunderlings, must be exposed to the internet.

## [ratking](https://github.com/firecow/ratking)
A rathole server manager, must be exposed to the internet, and have a range of ports avaiable to it.

## [ratunderling](https://github.com/firecow/ratunderling) 
A dataplane application creating ratholes and/or proxies based on command line options and council state


## Quickstart

```bash
# Start ratcouncil
docker run --rm --name ratcouncil -p 8080 firecow/ratcouncil
```

```bash
# Start ratking
docker run --rm --name ratking -p 2333 -p 5000-6000 firecow/ratking -e RATKING_PORT_RANGE=5000-6000 -e RATKING_HOSTNAME=localhost
```

```bash
# Create an network for ratunderling and arbitrary webserver to simulate a network not exposed to the internet
docker create network ratinners1
# Start ratunderling hole tells council that it wants a rathole called arbitrary, if traffic from "arbitrary" is received it will be sent to echoserver:8080
docker run --rm --name ratunderling-hole --network ratinners1 firecow/ratunderling --rathole "name=arbitrary to=echoserver:8080"
# Start an arbitrary webserver
docker run --rm --name echoserver --network ratinners1 jmalloc/echo-server:0.3.4
# Underling tells ratcouncil that it wants traffic from arbitray name directed to echoserver:8080
```

```bash
# Create an network for ratunderling-proxy and curl to simulate a different network location not exposed to the internet
docker create network ratinners2
# Start ratunderling proxy, creates proxy named arbitrary based on council state
docker run --rm --name ratunderling-proxy --network ratinners2 firecow/ratunderling --ratproxy "name=arbitrary listen=2183"
# Request echoserver via curl through the proxy
docker run --rm --name echoserver --network ratinners2 curlimages/curl curl --proxy http://underling-proxy:2183 http://whatnot
```
