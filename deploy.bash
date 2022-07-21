#!/usr/bin/env bash

image_name="nokal/waiting-time-backend"
container_name="waiting-time-backend"
remote_host=193.40.11.233
assets_dir=/home/ihar/deployments/waiting-time-backend/assets

ssh $remote_host docker pull $image_name
ssh $remote_host docker stop $container_name
ssh $remote_host docker rm $container_name
ssh $remote_host docker run -d -p 80:8080 -e WEBAPP_HOST=$remote_host -v $assets_dir:/srv/webapp/assets --name $container_name $image_name
