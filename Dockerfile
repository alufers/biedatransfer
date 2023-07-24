FROM --platform=$BUILDPLATFORM golang:alpine AS builder
ARG TARGETPLATFORM
ARG BUILDPLATFORM

# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git bash libmagic libmagic-static gcc alpine-sdk file-dev && mkdir -p /build/biedatransfer

WORKDIR /build/biedatransfer

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download -json

COPY . .

RUN mkdir -p /app && GOOS=${TARGETPLATFORM%%/*} GOARCH=${TARGETPLATFORM##*/} \
    go build -ldflags='-s -w' -o /app/biedatransfer .

# RUN echo "Running on architecture: $(uname -m), BUILDPLATFORM=$BUILDPLATFORM, TARGETPLATFORM=$TARGETPLATFORM" && exit 1

FROM alpine:edge

# add testing repository
RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

RUN apk update && apk add --no-cache libmagic exiftool binwalk ca-certificates 

COPY --from=builder /app/biedatransfer /app/biedatransfer

LABEL org.opencontainers.image.description A docker image for the biedatransfer telegram bot.

ENTRYPOINT ["/app/biedatransfer"]
