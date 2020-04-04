ARG GOLANG_VERSION=1.13-alpine
FROM golang:${GOLANG_VERSION} as builder
LABEL maintainer "pgillich ta gmail.com"

COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"'

# Making minimal image (only one binary)

#FROM scratch
FROM alpine

ARG RECEIVE_PORT="8088"

COPY --from=builder "/src/chat-bot" "/chat-bot"

EXPOSE ${RECEIVE_PORT}