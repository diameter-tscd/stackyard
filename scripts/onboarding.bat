@echo off
cls

setlocal EnableDelayedExpansion

:: Robustly switch to project root (one level up from this script)
cd /d "%~dp0.."

set "CONFIG_FILE=config.yaml"
set "BACKUP_FILE=config.yaml.backup"

:: Define ANSI Colors (limited support in Windows CMD)
:: Fancy Pastel Palette (matching TUI colors from pkg/tui/live.go)
set "RESET=[0m"
set "BOLD=[1m"
set "P_PURPLE=[38;5;219m"
set "B_PURPLE=[1;38;5;219m"
set "P_CYAN=[38;5;117m"
set "B_CYAN=[1;38;5;117m"
set "P_GREEN=[38;5;108m"
set "B_GREEN=[1;38;5;108m"
set "P_YELLOW=[93m"
set "B_YELLOW=[1;93m"
set "P_RED=[91m"
set "B_RED=[1;91m"
set "GRAY=[38;5;242m"
set "WHITE=[97m"
set "B_WHITE=[1;97m"

:: Check if build script exists
if not exist "scripts\build.bat" (
    echo Error: build.bat not found in scripts\ directory
    pause
    exit /b 1
)

:: Check if Go is installed
go version >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

:: Check Go version (minimum 1.24)
for /f "tokens=3" %%i in ('go version') do set go_ver=%%i
set go_ver=%go_ver:go=%
for /f "tokens=1,2 delims=." %%a in ("%go_ver%") do (
    set major=%%a
    set minor=%%b
)
if %major% gtr 1 goto version_ok
if %major% equ 1 if %minor% geq 24 goto version_ok
echo Error: Go version %go_ver% is too old. Minimum required is 1.24
echo Please install Go from https://golang.org/dl/
pause
exit /b 1
:version_ok

:: Function to read user input with default value
:read_input
setlocal EnableDelayedExpansion
set "prompt=%~1"
set "default=%~2"
set "input="

echo !prompt!
if "!default!" NEQ "" echo Default: !default!
set /p "input=Enter value: "
if "!input!"=="" if "!default!" NEQ "" set "input=!default!"
endlocal & set "input=%input%"
goto :eof

:: Function to read yes/no with default
:read_yes_no
setlocal EnableDelayedExpansion
set "prompt=%~1"
set "default=%~2"
set "input="

echo !prompt!
if "!default!"=="y" (
    echo Default: Yes
) else (
    echo Default: No
)
set /p "input=Choice (y/n): "
if "!input!"=="" set "input=!default!"

if /i "!input!"=="y" (
    echo true
) else if /i "!input!"=="yes" (
    echo true
) else (
    echo false
)
endlocal
goto :eof

:: Function to update config value (simple implementation)
:update_config
setlocal
set "key=%~1"
set "value=%~2"

:: Escape quotes in value
set "value=%value:"=""%"

:: Use CMD to update
if exist "%CONFIG_FILE%.tmp" del "%CONFIG_FILE%.tmp"
for /f "tokens=*" %%i in (%CONFIG_FILE%) do (
    echo %%i | findstr /b "%key%:" >nul 2>&1
    if !errorlevel! neq 0 (
        echo %%i >> "%CONFIG_FILE%.tmp"
    ) else (
        echo %key%: "%value%" >> "%CONFIG_FILE%.tmp"
    )
)
if exist "%CONFIG_FILE%.tmp" (
    move /y "%CONFIG_FILE%.tmp" "%CONFIG_FILE%" >nul 2>&1
    echo Configuration updated: %key%
) else (
    echo Warning: Could not update %key% in %CONFIG_FILE%
)
goto :eof

:: Function to update boolean config value
:update_config_bool
setlocal
set "key=%~1"
set "value=%~2"

:: Use CMD to update
if exist "%CONFIG_FILE%.tmp" del "%CONFIG_FILE%.tmp"
for /f "tokens=*" %%i in (%CONFIG_FILE%) do (
    echo %%i | findstr /b "%key%:" >nul 2>&1
    if !errorlevel! neq 0 (
        echo %%i >> "%CONFIG_FILE%.tmp"
    ) else (
        echo %key%: %value% >> "%CONFIG_FILE%.tmp"
    )
)
if exist "%CONFIG_FILE%.tmp" (
    move /y "%CONFIG_FILE%.tmp" "%CONFIG_FILE%" >nul 2>&1
    echo Configuration updated: %key%
) else (
    echo Warning: Could not update %key% in %CONFIG_FILE%
)
goto :eof

:: Function to show warning
:show_warning
echo.
echo WARNING: %~1
echo.
goto :eof

:: Function to show info
:show_info
echo.
echo INFO: %~1
echo.
goto :eof

:: Function to show success
:show_success
echo.
echo SUCCESS: %~1
echo.
goto :eof

:: Check if config.yaml exists
if not exist "%CONFIG_FILE%" (
    echo Error: %CONFIG_FILE% not found in current directory
    echo Please run this script from the project root directory.
    pause
    exit /b 1
)

:: Backup original config
copy "%CONFIG_FILE%" "%BACKUP_FILE%" >nul

echo.
echo    /\ 
echo   (  )   stackyard Onboarding by diameter-tscd
echo    \/
echo ------------------------------------------------------------------------------
echo Welcome to the stackyard onboarding setup!
echo This script will help you configure your application.
echo ------------------------------------------------------------------------------

:: Basic Application Configuration
echo BASIC APPLICATION CONFIGURATION

call :read_input "Enter application name" "My Fancy Go App"
set "APP_NAME=%errorlevel%"%

call :read_input "Enter application version" "1.0.0"
set "APP_VERSION=%errorlevel%"%

call :read_input "Enter server port" "8080"
set "SERVER_PORT=%errorlevel%"%

call :read_input "Enter monitoring port" "9090"
set "MONITORING_PORT=%errorlevel%"%

echo ENVIRONMENT SETTINGS

call :read_yes_no "Enable debug mode?" "y"
set "DEBUG_MODE=%errorlevel%"%

call :read_yes_no "Enable TUI (Terminal User Interface)?" "y"
set "TUI_MODE=%errorlevel%"%

call :read_yes_no "Quiet startup (suppress console logs)?" "n"
set "QUIET_STARTUP=%errorlevel%"%

echo SERVICE CONFIGURATION

call :read_yes_no "Enable monitoring dashboard?" "y"
set "ENABLE_MONITORING=%errorlevel%"%

call :read_yes_no "Enable API encryption?" "n"
set "ENABLE_ENCRYPTION=%errorlevel%"%

echo INFRASTRUCTURE CONFIGURATION

call :read_yes_no "Enable Redis?" "n"
set "ENABLE_REDIS=%errorlevel%"%

set /p "ENABLE_POSTGRES=Enable PostgreSQL? (single/multi/none) [single]: "
if "%ENABLE_POSTGRES%"=="" set "ENABLE_POSTGRES=single"

call :read_yes_no "Enable Kafka?" "n"
set "ENABLE_KAFKA=%errorlevel%"%

call :read_yes_no "Enable MinIO (Object Storage)?" "n"
set "ENABLE_MINIO=%errorlevel%"%

echo APPLYING CONFIGURATION

:: Update basic config
call :update_config "app.name" "%APP_NAME%"
call :update_config "app.version" "%APP_VERSION%"
call :update_config "server.port" "%SERVER_PORT%"
call :update_config "monitoring.port" "%MONITORING_PORT%"

:: Update boolean configs
call :update_config_bool "app.debug" "%DEBUG_MODE%"
call :update_config_bool "app.enable_tui" "%TUI_MODE%"
call :update_config_bool "app.quiet_startup" "%QUIET_STARTUP%"
call :update_config_bool "monitoring.enabled" "%ENABLE_MONITORING%"
call :update_config_bool "encryption.enabled" "%ENABLE_ENCRYPTION%"
call :update_config_bool "redis.enabled" "%ENABLE_REDIS%"
call :update_config_bool "kafka.enabled" "%ENABLE_KAFKA%"

:: Handle PostgreSQL configuration
if "%ENABLE_POSTGRES%"=="single" (
    call :update_config_bool "postgres.enabled" "true"
) else if "%ENABLE_POSTGRES%"=="multi" (
    call :update_config_bool "postgres.enabled" "true"
) else (
    call :update_config_bool "postgres.enabled" "false"
)

:: Handle MinIO
call :update_config_bool "monitoring.minio.enabled" "%ENABLE_MINIO%"

call :show_success "Configuration updated successfully!"

:: Security Warnings
echo.
echo SECURITY WARNINGS

call :show_warning "Default credentials are configured. You MUST change these before production use:"
echo • PostgreSQL password: 'Mypostgres01'
echo • Monitoring password: 'admin'
echo • MinIO credentials: 'minioadmin/minioadmin'
echo • API secret key: 'super-secret-key'

call :show_warning "API obfuscation is enabled. This adds security through obscurity but is not encryption."

if "%ENABLE_ENCRYPTION%"=="true" (
    call :show_warning "Encryption is enabled but no key is set. You need to configure 'encryption.key' in config.yaml"
)

:: Next Steps
echo.
echo NEXT STEPS

call :show_info "1. Review and customize config.yaml with your specific settings"
call :show_info "2. Update all default passwords and secrets"
call :show_info "3. Set up your infrastructure (PostgreSQL, Redis, etc.)"
call :show_info "4. Run 'go mod tidy' to ensure dependencies are correct"
call :show_info "5. Build the application using 'scripts\build.bat'"
call :show_info "6. Test the application with 'go run cmd\app\main.go'"

:: Offer to run additional setup
echo.
echo Would you like to run additional setup commands?
echo This will run 'go mod tidy' and check for build issues.
call :read_yes_no "Run setup commands?" "y"
set "RUN_SETUP=%errorlevel%"%

if "%RUN_SETUP%"=="true" (
    echo.
    echo RUNNING SETUP COMMANDS

    echo.
    echo Running 'go mod tidy'...
    go mod tidy
    if %ERRORLEVEL% EQU 0 (
        call :show_success "Dependencies updated successfully"
    ) else (
        call :show_warning "Failed to update dependencies - you may need to check your Go installation"
    )

    echo.
    echo Checking build...
    go build -o temp_build.exe .\cmd\app\main.go
    if %ERRORLEVEL% EQU 0 (
        call :show_success "Build test successful"
        if exist "temp_build.exe" del "temp_build.exe"
    ) else (
        call :show_warning "Build failed - check your configuration and dependencies"
    )

    echo.
)

:: Final message
echo ======================================================================
echo   ONBOARDING COMPLETE! Your app is ready to go!
echo ======================================================================
echo.
echo Backup created: %BACKUP_FILE%
echo Configuration: %CONFIG_FILE%
echo.
echo Happy coding!
echo.

pause
endlocal
