#!/usr/bin/env bash

remote_host=193.40.11.233
deployment_dir=/home/ihar/deployments/waiting-time-backend
file_option="--file $deployment_dir/docker-compose.yml"

printf "\n🤞 Copying configuration files\n"
scp docker-compose.yml $remote_host:$deployment_dir/
scp nginx.conf $remote_host:$deployment_dir/
scp env.production $remote_host:$deployment_dir/.env

printf "\n🤞 Pulling updated Docker images\n"
ssh $remote_host docker compose $file_option pull

printf "\n🤞 Stopping the current deployment\n"
ssh $remote_host docker compose $file_option down

printf "\n🤞 Starting the new deployment\n"
ssh $remote_host docker compose $file_option up --no-build --remove-orphans --detach