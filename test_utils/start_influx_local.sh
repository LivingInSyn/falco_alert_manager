docker run --rm -e DOCKER_INFLUXDB_INIT_USERNAME=fam_root \
-e DOCKER_INFLUXDB_INIT_PASSWORD=fam_never_in_production_please \
-e DOCKER_INFLUXDB_INIT_ORG=falco_alert_manager \
-e DOCKER_INFLUXDB_INIT_BUCKET=events \
-e DOCKER_INFLUXDB_INIT_RETENTION=1w \
-e DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=fam_token_never_in_production_please \
-e DOCKER_INFLUXDB_INIT_MODE=setup \
-e INFLUXD_BOLT_PATH=/home/influxdb/.influxdbv2/influxd.bolt \
influxdb:2.3