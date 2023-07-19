FROM golang:buster AS build
COPY . /build
COPY .git /build/.git
WORKDIR /build
RUN go build -o target/ ./cmd/avalond/...

FROM ubuntu:20.04
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=build /build/target/avalond /usr/bin/avalond
ENTRYPOINT /usr/bin/avalond
