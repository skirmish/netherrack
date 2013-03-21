@echo off
set GOPATH=%CD%
go install Netherrack
bin\Netherrack.exe -debugnames
pause
