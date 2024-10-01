run-grafana:
	docker run -d \
    --name=grafana01 \
    --restart=always \
    --net=prometheus_prom_net \
    -p 3100:3000 \
    docker.io/grafana/grafana-oss:9.5.6

add-exporter:
	docker cp $1 prometheus_prometheus_1:/prometheus