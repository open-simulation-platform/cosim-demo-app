@echo off
start .\bin\cse-server-go.exe
timeout 2 > nul
start http://localhost:8000
