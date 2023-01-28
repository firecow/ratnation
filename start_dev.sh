#!/usr/bin/env bash
set -eo pipefail

trap "docker rm -f ratnation-echo-server" SIGINT

npx nodemon src/index.js council &

npx nodemon src/index.js king \
  --host="$(hostname -f)" \
  --rathole="bind_port=2334 ports=5000-5001" &

npx nodemon src/index.js ling \
  --ling_id="0a976e7a-87c5-4549-9431-e4881c740cec" \
  --rathole="name=alpha local_addr=localhost:3000" \
  --proxy="name=alpha bind_port=2183" &

npx nodemon src/index.js ling \
  --ling_id="0573442b-5491-444e-9c63-c2907079ff5f" \
  --rathole="name=alpha local_addr=localhost:3000" \
  --proxy="name=alpha bind_port=2184" &

docker run --rm --name=ratnation-echo-server -p 3000:8080 jmalloc/echo-server &

wait
