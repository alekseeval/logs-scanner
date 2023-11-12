# syntax=docker/dockerfile:1

# BUILD stage
FROM golang:1.21.3-alpine as build-image

WORKDIR /etc/scanner/

COPY . ./
RUN go mod download
RUN go build -o ./scanner /etc/scanner/cmd/main.go

# DEPLOY stage
FROM alpine:3.15

WORKDIR /etc/scanner

COPY --from=build-image /etc/scanner/scanner ./
COPY --from=build-image /etc/scanner/config.json ./
COPY --from=build-image /etc/scanner/static ./static/

CMD [ "/etc/scanner/scanner" ]