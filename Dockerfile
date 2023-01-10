FROM alpine:3.17.0 as node_modules
RUN apk add nodejs npm
COPY package.json package-lock.json ./
RUN npm install --no-audit --no-progress --omit=dev

FROM alpine:3.17.0

RUN apk add nodejs socat && \
 wget -O rathole.zip https://github.com/rapiz1/rathole/releases/download/v0.4.7/rathole-x86_64-unknown-linux-musl.zip && \
 unzip rathole.zip && rm -f rathole.zip && \
 mv rathole /usr/bin/rathole

COPY --from=node_modules /node_modules /node_modules
COPY package.json ./
COPY src src
ENTRYPOINT ["node", "src/index.js"]
