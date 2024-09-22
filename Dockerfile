FROM docker.io/golang:1.22 AS builder
RUN mkdir /app
WORKDIR /app
COPY . /app
RUN make build-linux

FROM docker.io/alpine:latest
RUN mkdir /app && adduser -h /app -D restapi
WORKDIR /app
COPY --chown=restapi --from=builder /app/resapi .
EXPOSE 4000
CMD ["/app/resapi"]