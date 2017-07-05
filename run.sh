#!/bin/bash
cd src; go build -o anomaly_detector; cd ..
src/anomaly_detector ./log_input/batch_log.json ./log_input/stream_log.json ./log_output/flagged_purchases.json
