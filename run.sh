#!/bin/bash
cd src; go build -o blah; cd ..
src/blah ./log_input/batch_log.json ./log_input/stream_log.json ./log_output/flagged_purchases.json
