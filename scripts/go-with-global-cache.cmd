@echo off
setlocal

if "%~1"=="" (
  echo Usage: scripts\go-with-global-cache.cmd ^<go arguments^>
  exit /b 1
)

if defined LOCALAPPDATA (
  set "GOCACHE=%LOCALAPPDATA%\go-build"
) else if defined USERPROFILE (
  set "GOCACHE=%USERPROFILE%\AppData\Local\go-build"
) else (
  set "GOCACHE=%CD%\.cache\go-build"
)

if not exist "%GOCACHE%" mkdir "%GOCACHE%"

go %*
exit /b %ERRORLEVEL%
