#!/usr/bin/env bash

docker run -it -p 8080:8080 -e WEBAPP_HOST=localhost:8080 nokal/waiting-time-backend