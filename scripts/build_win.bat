@echo off
cd ..
go build -ldflags="-s -w" -o .\bin\app.exe main.go
cd scripts
dir .\..\bin