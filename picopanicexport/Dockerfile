FROM docker.io/golang:1.23 AS builder
RUN mkdir /app
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o panicexport -ldflags="-s -w"

FROM docker.io/alpine:latest
RUN mkdir /app && adduser -h /app -D panicexport
WORKDIR /app
COPY --chown=panicexport --from=builder /app/panicexport .
EXPOSE 3030
CMD ["/app/panicexport"]