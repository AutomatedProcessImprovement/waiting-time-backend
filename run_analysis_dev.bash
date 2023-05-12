#!/usr/bin/env bash

cd /usr/src/app
poetry run wta --log_path "$1" --output_dir "$2"
