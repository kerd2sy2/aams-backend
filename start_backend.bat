@echo off
cd /d "d:\aams\backend"

echo Cleaning port 8080...
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :8080') do taskkill /f /pid %%a 2>nul
taskkill /f /im api.exe 2>nul

set PORT=8080
set GIN_MODE=release
set JWT_SECRET=hV6bI9UDaAw1zqlMdTL7tEB3orZ5gQpPfJCGnFm0N8vcixesHOkyRK4uWS2YXj
set JWT_REFRESH_SECRET=g3SwHtEdyKFrViDNkXIuWb92xhTq81YAvpRB4Gfo7nUl0a6JezCjLcm5OQMsPZ

echo Starting Backend...
if exist api.exe (
    start /b api.exe
) else (
    start /b go run .\cmd\api
)
