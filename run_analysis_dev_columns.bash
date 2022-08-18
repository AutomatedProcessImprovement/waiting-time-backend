#!/usr/bin/env bash

source /Users/ihar/Projects/PIX/waiting-time-analysis/venv/bin/activate
export RSCRIPT_BIN_PATH=/usr/local/bin/Rscript
wta --log_path $1 --output_dir $2 --columns_json $3
