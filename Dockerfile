FROM traefik:v3.3.4 AS traefik

FROM alpine:3.21.3 AS rathole
RUN wget -O rathole.zip https://github.com/rapiz1/rathole/releases/download/v0.4.8/rathole-x86_64-unknown-linux-musl.zip && unzip rathole.zip

FROM alpine:3.21.3 AS node_modules
RUN apk add nodejs npm
COPY package.json package-lock.json ./
RUN npm install --no-audit --no-progress --omit=dev

FROM alpine:3.21.3
RUN apk add nodejs npm
COPY --from=node_modules /node_modules /node_modules
COPY --from=traefik /usr/local/bin/traefik /usr/local/bin/traefik
COPY --from=rathole /rathole /usr/local/bin/rathole
COPY package.json ./
COPY src src
ENTRYPOINT ["node", "src/index.js"]
