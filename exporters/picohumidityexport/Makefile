build:
	docker build -t picohumidity:v1 .

run:
	docker run -d \
    --name picohumidity-v1 \
    -p 3030:3030 \
    --env PICO_SERVER_URL=http://192.168.2.137 \
    --restart=always \
    --net=prometheus_prom_net \
    picohumidity:v1