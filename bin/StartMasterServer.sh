#!/bin/bash
chmod +x ./master_server
mkdir -p ./log
nohup ./master_server > ./log/master_server.stdout 2>&1 &
