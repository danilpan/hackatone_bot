#!/bin/sh -e

env GOOS=linux GOARCH=amd64 go build -o app -buildvcs=false