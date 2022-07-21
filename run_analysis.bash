#!/usr/bin/env bash

source /usr/src/app/venv/bin/activate
export RSCRIPT_BIN_PATH=/usr/bin/Rscript
wta --log_path $1 --output_dir $2
