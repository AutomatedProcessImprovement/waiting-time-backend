#!/usr/bin/env bash

cp env.development .env
docker compose --file docker-compose.yml up --remove-orphans --build
