#!/bin/bash

cd ..
go build -ldflags="-s -w" -o ./bin/app main.go
chmod +x ./bin/app
cd scripts
ls ./../bin