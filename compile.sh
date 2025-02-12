#!/bin/bash

set -x

#GOOS=linux GOARCH=mipsle go build -ldflags="-s -w" -o intercom_mipsle main.go
#GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o intercom_amd64 main.go
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o intercom_arm64 main.go
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o intercom_armv7 main.go
##upx

if [[ "$1" == "upx" ]]; then

UPX_BINARY=upx
#$UPX_BINARY -9 --ultra-brute intercom_mipsle
#$UPX_BINARY -9 --ultra-brute intercom_amd64
$UPX_BINARY -9 --ultra-brute intercom_arm64
$UPX_BINARY -9 --ultra-brute intercom_armv7
fi
