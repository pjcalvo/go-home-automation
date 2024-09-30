build:
	docker build -t panicexport:v1 .

run:
	docker run -d \
    --name panicexport-v1 \
    -p 3030:3030 \
    --env PICO_SERVER_URL=http://192.168.2.137 \
    --restart=always \
    --net=prometheus_prom_net \
    panicexport:v1