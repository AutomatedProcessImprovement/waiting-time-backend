#!/usr/bin/env bash

docker run -it -p 8080:8080 -e WEBAPP_HOST=193.40.11.233 nokal/waiting-time-backend
