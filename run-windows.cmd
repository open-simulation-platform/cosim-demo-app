@echo off
cd bin
start cosim-demo-app.exe
timeout 2 > nul
start http://localhost:8000
