#!/usr/bin/env bash

source /usr/src/app/venv/bin/activate
wta --log_path $1 --output_dir $2 --columns_json $3
