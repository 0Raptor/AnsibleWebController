@echo off
set GOARCH=amd64
set GOOS=linux
rem go tool dist install -v pkg/runtime
rem go install -v -a std
go build
pause