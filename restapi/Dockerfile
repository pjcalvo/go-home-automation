FROM docker.io/golang:1.23 AS builder
RUN mkdir /app
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o restapi -ldflags="-s -w"
# RUN go build -o restapi

FROM docker.io/alpine:latest AS runner
RUN mkdir /app && adduser -h /app -D restapi
WORKDIR /app
COPY --chown=restapi --from=builder /app/restapi .
EXPOSE 4000
CMD ["/app/restapi"]