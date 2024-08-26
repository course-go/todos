FROM golang:alpine AS builder

WORKDIR /usr/app

RUN apk update \
    && apk --no-cache --update add build-base

COPY . .
RUN make build

FROM alpine:latest

WORKDIR /usr/app

COPY --from=builder /usr/app/bin/todos /usr/app/todos

CMD ["/usr/app/todos"]
