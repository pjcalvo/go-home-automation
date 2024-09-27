build:
	docker build -t picotempexport:v1 .

run:
	docker run -d \
    --name picotempexport-v1 \
    -p 3030:3030 \
    --env PICO_SERVER_URL=http://192.168.2.137 \
    --restart=always \
    --net=prometheus_prom_net \
    picotempexport:v1