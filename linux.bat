@echo off
set GOPATH=%CD%
set GOOS=linux
set GOARCH=386
set CGO_ENABLED=0
go install Karascraft
pause