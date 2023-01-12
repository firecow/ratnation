#!/usr/bin/env bash

set -e

npx nodemon src/index.js council &

npx nodemon src/index.js king \
  --host="$(hostname -f)" \
  --rathole="bind_port=2334 ports=5000-5009" \
  &
npx nodemon src/index.js ling \
  --rathole="name=alpha local_addr=localhost:3000" \
  --socat="name=alpha bind_port=2189" &

docker run --rm -p 3000:8080 jmalloc/echo-server &

wait
