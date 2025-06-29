ARG GOLANG_VERSION
ARG ALPINE_VERSION
FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk --no-cache --virtual .build-deps add make gcc musl-dev binutils-gold

COPY . /app
WORKDIR /app

RUN make build

FROM alpine:${ALPINE_VERSION}

RUN apk upgrade --no-cache --no-interactive && apk add --no-cache ca-certificates tzdata && \
  adduser -u 1000 -S -D -H krakend && \
  mkdir /etc/goodin


COPY --from=builder /app/goodin /usr/bin/goodin
COPY config.docker.toml /etc/tiultemplate/config.toml
COPY /migrations    /etc/goodin/migrations

RUN chown 1000 /etc/goodin

USER 1000

WORKDIR /etc/goodin

ENTRYPOINT [ "/usr/bin/goodin" ]
CMD [ "all" ]

EXPOSE 3999
