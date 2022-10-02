FROM golang:1.19.1-alpine as build-env
RUN go install -v github.com/projectdiscovery/katana/cmd/katana@latest

FROM alpine:3.16.2
RUN apk add --no-cache bind-tools ca-certificates chromium
COPY --from=build-env /go/bin/katana /usr/local/bin/katana
ENTRYPOINT ["katana"]