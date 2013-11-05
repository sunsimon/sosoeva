#!/bin/bash
chmod +x ./server
mkdir -p ./log
nohup ./server > ./log/server.stdout 2>&1 &
