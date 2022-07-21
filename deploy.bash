#!/usr/bin/env bash

image_name="nokal/waiting-time-backend"
swagger_image_name="swaggerapi/swagger-ui"
#container_name="waiting-time-backend"
remote_host=193.40.11.233
deployment_dir=/home/ihar/deployments/waiting-time-backend
#assets_dir=$deployment_dir/assets

#ssh $remote_host docker pull $image_name
#ssh $remote_host docker stop $container_name
#ssh $remote_host docker rm $container_name
#ssh $remote_host docker run -d -p 80:8080 -e WEBAPP_HOST=$remote_host -v $assets_dir:/srv/webapp/assets --name $container_name $image_name

ssh $remote_host docker pull $image_name
ssh $remote_host docker pull $swagger_image_name

scp docker-compose.yml $remote_host:$deployment_dir/
scp nginx.conf $remote_host:$deployment_dir/
ssh $remote_host cd $deployment_dir && docker compose down
ssh $remote_host cd $deployment_dir && WEBAPP_HOST=$remote_host docker compose up --remove-orphans -d --wait