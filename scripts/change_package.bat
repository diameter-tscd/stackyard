@echo off
setlocal enabledelayedexpansion

:: Define ANSI Colors
set "RESET=[0m"
set "BOLD=[1m"
set "DIM=[2m"
set "UNDERLINE=[4m"

:: Fancy Pastel Palette (main color: #8daea5)
set "P_PURPLE=[38;5;108m"
set "B_PURPLE=[1;38;5;108m"
set "P_CYAN=[38;5;117m"
set "B_CYAN=[1;38;5;117m"
set "P_GREEN=[38;5;46m"
set "B_GREEN=[1;38;5;46m"
set "P_YELLOW=[93m"
set "B_YELLOW=[1;93m"
set "P_RED=[91m"
set "B_RED=[1;91m"
set "GRAY=[38;5;242m"
set "WHITE=[97m"
set "B_WHITE=[1;97m"

REM Script to change the Go module package name
REM Usage: change_package.bat <new_module_name>

if "%~1"=="" (
  echo Usage: %0 ^<new_module_name^>
  echo Example: %0 github.com/user/new-project
  exit /b 1
)

set NEW_MODULE=%~1

REM Get current module name from go.mod
for /f "tokens=2" %%i in ('findstr "^module " go.mod') do set CURRENT_MODULE=%%i

if "%CURRENT_MODULE%"=="" (
  echo Error: Could not find module declaration in go.mod
  exit /b 1
)

echo Changing module from '%CURRENT_MODULE%' to '%NEW_MODULE%'

REM Change module name in go.mod
set "search=module %CURRENT_MODULE%"
set "replace=module %NEW_MODULE%"
set "tempfile=%temp%\go_mod_temp.txt"
(for /f "delims=" %%i in (go.mod) do (
  set "line=%%i"
  set "line=!line:%search%=%replace%!"
  echo !line!
)) > "%tempfile%"
move "%tempfile%" go.mod >nul

if %ERRORLEVEL% neq 0 (
  echo Error: Failed to update go.mod
  exit /b 1
)

REM Update all import paths in .go files
for /r %%f in (*.go) do (
  set "tempfile=%%f.tmp"
  (for /f "delims=" %%i in ("%%f") do (
    set "line=%%i"
    set "line=!line:%CURRENT_MODULE%=%NEW_MODULE%!"
    echo !line!
  )) > "!tempfile!"
  move "!tempfile!" "%%f" >nul
)

echo Successfully changed module name and updated imports
echo Note: No backup files created in this version. Make sure to commit changes before running.
endlocal
