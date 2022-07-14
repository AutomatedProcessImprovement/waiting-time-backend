#!/usr/bin/env bash

remote_host=193.40.11.233
results_dir=/home/ihar/deployments/waiting-time-backend/results

ssh $remote_host docker pull nokal/waiting-time-backend
ssh $remote_host docker stop waiting-time-backend
ssh $remote_host docker rm waiting-time-backend
ssh $remote_host docker run -d -p 80:8080 -e WEBAPP_HOST=$remote_host -v $results_dir:/srv/webapp/assets/results --name waiting-time-backend nokal/waiting-time-backend
