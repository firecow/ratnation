FROM traefik:v2.10.4 AS traefik

FROM alpine:3.18.2 AS node_modules
RUN apk add nodejs npm
COPY package.json package-lock.json ./
RUN npm install --no-audit --no-progress --omit=dev

FROM alpine:3.18.2
RUN apk add nodejs curl && \
 wget -O rathole.zip https://github.com/rapiz1/rathole/releases/download/v0.4.7/rathole-x86_64-unknown-linux-musl.zip && \
 unzip rathole.zip && rm -f rathole.zip && \
 mv rathole /usr/bin/rathole
COPY --from=node_modules /node_modules /node_modules
COPY --from=traefik /usr/local/bin/traefik /usr/local/bin/traefik
COPY package.json ./
COPY src src
ENTRYPOINT ["node", "src/index.mjs"]
