FROM golang:buster AS build
COPY . /build
WORKDIR /build
RUN go build -o target/ ./cmd/avalond/...

FROM ubuntu:20.04
COPY --from=build /build/target/avalond /usr/bin/avalond
ENTRYPOINT /usr/bin/avalond
