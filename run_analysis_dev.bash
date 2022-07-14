#!/usr/bin/env bash

source /Users/ihar/Projects/PIX/process-waste/venv/bin/activate
export RSCRIPT_BIN_PATH=/usr/local/bin/Rscript
process-waste --log_path $1 --output_dir $2
