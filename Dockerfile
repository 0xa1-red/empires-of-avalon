FROM golang:bullseye AS build
COPY . /build
COPY .git /build/.git
WORKDIR /build
RUN make build

FROM ubuntu:20.04
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=build /build/target/avalond /usr/bin/avalond
ENTRYPOINT /usr/bin/avalond
