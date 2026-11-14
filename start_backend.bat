@echo off
cd /d "d:\aams\backend"

echo Cleaning port 8080...
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :8080') do taskkill /f /pid %%a 2>nul
taskkill /f /im api.exe 2>nul

set PORT=8080
set GIN_MODE=release
rem Secrets (JWT_SECRET, JWT_REFRESH_SECRET) are loaded from .env by config.LoadConfig()

echo Starting Backend...
if exist api.exe (
    start /b api.exe
) else (
    start /b go run .\cmd\api
)
