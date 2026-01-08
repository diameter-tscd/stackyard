@echo off
cls

setlocal EnableDelayedExpansion
set "DIST_DIR=dist"
set "APP_NAME=stackyard.exe"
set "MAIN_PATH=./cmd/app/main.go"

:: Define ANSI Colors
for /F "tokens=1,2 delims=#" %%a in ('"prompt #$H#$E# & echo on & for %%b in (1) do rem"') do (
  set "ESC=%%b"
)
set "RESET=%ESC%[0m"
set "BOLD=%ESC%[1m"
set "DIM=%ESC%[2m"
set "UNDERLINE=%ESC%[4m"

:: Fancy Pastel Palette (main color: #8daea5)
set "P_PURPLE=%ESC%[38;5;108m"
set "B_PURPLE=%ESC%[1;38;5;108m"
set "P_CYAN=%ESC%[38;5;117m"
set "B_CYAN=%ESC%[1;38;5;117m"
set "P_GREEN=%ESC%[38;5;46m"
set "B_GREEN=%ESC%[1;38;5;46m"
set "P_YELLOW=%ESC%[93m"
set "B_YELLOW=%ESC%[1;93m"
set "P_RED=%ESC%[91m"
set "B_RED=%ESC%[1;91m"
set "GRAY=%ESC%[38;5;242m"
set "WHITE=%ESC%[97m"
set "B_WHITE=%ESC%[1;97m"

:: Robustly switch to project root
cd /d "%~dp0.."

echo.
echo    %P_PURPLE%(\_/)%RESET%
echo    %P_PURPLE%(o.o)%RESET%   %B_PURPLE%%APP_NAME% Builder%RESET% %GRAY%by%RESET% %B_WHITE%diameter-tscd%RESET%
echo   %P_PURPLE%c(")(")%RESET%
echo %GRAY%----------------------------------------------------------------------%RESET%

REM 0. Check required tools
echo %B_PURPLE%[0/6]%RESET% %P_CYAN%Checking required tools...%RESET%

REM Check goversioninfo
where goversioninfo >nul 2>nul
if %errorlevel% neq 0 (
    echo    %B_YELLOW%! goversioninfo not found. Installing...%RESET%
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
    if errorlevel 1 (
        echo    %B_RED%x Failed to install goversioninfo%RESET%
        exit /b 1
    )
    echo    %B_GREEN%+ goversioninfo installed%RESET%
) else (
    echo    %B_GREEN%+ goversioninfo found%RESET%
)

REM Check garble
where garble >nul 2>nul
if %errorlevel% neq 0 (
    echo    %B_YELLOW%! garble not found. Installing...%RESET%
    go install mvdan.cc/garble@latest
    if errorlevel 1 (
        echo    %B_RED%x Failed to install garble%RESET%
        exit /b 1
    )
    echo    %B_GREEN%+ garble installed%RESET%
) else (
    echo    %B_GREEN%+ garble found%RESET%
)

REM Ask user about garble build
echo %B_YELLOW%Use garble build for obfuscation? (Y/N, default N, timeout 10s): %RESET%
choice /T 10 /D N /C YN /N
if %errorlevel% equ 1 (
    set "USE_GARBLE=true"
    echo    %B_GREEN%+ Using garble build%RESET%
) else (
    set "USE_GARBLE=false"
    echo    %B_CYAN%+ Using regular go build%RESET%
)

REM 1. Generate Timestamp
set "TIMESTAMP=%date:~-4%%date:~4,2%%date:~7,2%_%time:~0,2%%time:~3,2%%time:~6,2%"
set "TIMESTAMP=%TIMESTAMP: =0%"
set "TIMESTAMP=%TIMESTAMP::=%"
set "TIMESTAMP=%TIMESTAMP:/=%"
set "BACKUP_ROOT=%DIST_DIR%\backups"
set "BACKUP_PATH=%BACKUP_ROOT%\%TIMESTAMP%"

REM 2. Stop running process
echo %B_PURPLE%[1/6]%RESET% %P_CYAN%Checking for running process...%RESET%
tasklist /FI "IMAGENAME eq %APP_NAME%" 2>NUL | find /I /N "%APP_NAME%">NUL
if "%ERRORLEVEL%"=="0" (
    echo    %B_YELLOW%! App is running. Stopping...%RESET%
    taskkill /F /IM %APP_NAME% >NUL
    timeout /t 1 /nobreak >NUL
) else (
    echo    %B_GREEN%+ App is not running.%RESET%
)

REM 3. Backup Old Files
echo %B_PURPLE%[3/6]%RESET% %P_CYAN%Backing up old files...%RESET%
if exist "%DIST_DIR%" (
    if not exist "%BACKUP_PATH%" mkdir "%BACKUP_PATH%"
    
    if exist "%DIST_DIR%\%APP_NAME%" (
        echo    %GRAY%- Moving old binary...%RESET%
        move "%DIST_DIR%\%APP_NAME%" "%BACKUP_PATH%\" >NUL
    )
    if exist "%DIST_DIR%\config.yaml" (
        move "%DIST_DIR%\config.yaml" "%BACKUP_PATH%\" >NUL
    )
    if exist "%DIST_DIR%\banner.txt" (
        move "%DIST_DIR%\banner.txt" "%BACKUP_PATH%\" >NUL
    )
    if exist "%DIST_DIR%\monitoring_users.db" (
        echo    %GRAY%- Backing up database...%RESET%
        move "%DIST_DIR%\monitoring_users.db" "%BACKUP_PATH%\" >NUL
    )
    if exist "%DIST_DIR%\web" (
        echo    %GRAY%- Moving old web assets...%RESET%
        move "%DIST_DIR%\web" "%BACKUP_PATH%\" >NUL
    )
    
    echo    %B_GREEN%+ Backup created at:%RESET% %B_WHITE%%BACKUP_PATH%%RESET%
) else (
    echo    %GRAY%+ No existing dist directory. Skipping backup.%RESET%
    mkdir "%DIST_DIR%"
)

REM 6. Archive Backup
echo %B_PURPLE%[4/6]%RESET% %P_CYAN%Archiving backup...%RESET%
if exist "%BACKUP_PATH%" (
    pushd "%BACKUP_ROOT%"
    tar -acf "%TIMESTAMP%.zip" "%TIMESTAMP%" 2>NUL
    popd
    if exist "%BACKUP_PATH%" rmdir /s /q "%BACKUP_PATH%"
    echo    %B_GREEN%+ Backup archived:%RESET% %B_WHITE%%BACKUP_ROOT%\%TIMESTAMP%.zip%RESET%
) else (
    echo    %GRAY%+ No backup created. Skipping archive.%RESET%
)

REM Ensure dist directory
if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"

REM 4. Build
echo %B_PURPLE%[5/6]%RESET% %P_CYAN%Building Go binary...%RESET%
goversioninfo -platform-specific
if "%USE_GARBLE%"=="true" (
    garble build -ldflags="-s -w" -o "%DIST_DIR%\%APP_NAME%" %MAIN_PATH%
) else (
    go build -ldflags="-s -w" -o "%DIST_DIR%\%APP_NAME%" %MAIN_PATH%
)
if %ERRORLEVEL% NEQ 0 (
    echo    %B_RED%x Build FAILED! Exit code: %ERRORLEVEL%%RESET%
    exit /b %ERRORLEVEL%
)
echo    %B_GREEN%+ Build successful:%RESET% %B_WHITE%%DIST_DIR%\%APP_NAME%%RESET%

REM 5. Copy Assets
echo %B_PURPLE%[6/6]%RESET% %P_CYAN%Copying assets...%RESET%

if exist "web" (
    echo    %B_GREEN%+ Copying web folder...%RESET%
    xcopy /E /I /Y /Q "web" "%DIST_DIR%\web" >NUL
)

if exist "config.yaml" (
    echo    %B_GREEN%+ Copying config.yaml...%RESET%
    copy /Y "config.yaml" "%DIST_DIR%" >NUL
)

if exist "banner.txt" (
    echo    %B_GREEN%+ Copying banner.txt...%RESET%
    copy /Y "banner.txt" "%DIST_DIR%" >NUL
)

if exist "monitoring_users.db" (
    echo    %B_GREEN%+ Copying monitoring_users.db...%RESET%
    copy /Y "monitoring_users.db" "%DIST_DIR%" >NUL
)

echo.
echo %GRAY%======================================================================%RESET%
echo  %B_PURPLE%SUCCESS!%RESET% %P_GREEN%Build ready at:%RESET% %UNDERLINE%%B_WHITE%%DIST_DIR%\%RESET%
echo %GRAY%======================================================================%RESET%
endlocal
