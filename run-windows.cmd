@echo off
start .\bin\cosim-demo-app.exe
timeout 2 > nul
start http://localhost:8000
