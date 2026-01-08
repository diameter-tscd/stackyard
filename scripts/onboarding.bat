@echo off
cls

:: Check if build script exists
if not exist "scripts\build.bat" (
    echo %B_RED%Error: build.bat not found in scripts\ directory%RESET%
    pause
    exit /b 1
)

:: Check if Go is installed
go version >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo %B_RED%Error: Go is not installed or not in PATH%RESET%
    echo %WHITE%Please install Go from https://golang.org/dl/%RESET%
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
echo %B_RED%Error: Go version %go_ver% is too old. Minimum required is 1.24%RESET%
echo %WHITE%Please upgrade Go from https://golang.org/dl/%RESET%
pause
exit /b 1
:version_ok

setlocal EnableDelayedExpansion
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

:: Function to read user input with default value
:read_input
setlocal
set "prompt=%~1"
set "default=%~2"
set "input="

echo.
echo %P_CYAN%%prompt%%RESET%
if "%default%" NEQ "" echo %GRAY%Default: %default%%RESET%
set /p "input=%B_WHITE%> %RESET%"
if "%input%"=="" if "%default%" NEQ "" set "input=%default%"
echo %input%
goto :eof

:: Function to read yes/no with default
:read_yes_no
setlocal
set "prompt=%~1"
set "default=%~2"
set "input="

echo.
echo %P_CYAN%%prompt%%RESET%
if "%default%"=="y" (
    echo %GRAY%Default: Yes%RESET%
) else (
    echo %GRAY%Default: No%RESET%
)
set /p "input=%B_WHITE%(y/n) > %RESET%"
if "%input%"=="" set "input=%default%"

if /i "%input%"=="y" (
    echo true
) else if /i "%input%"=="yes" (
    echo true
) else (
    echo false
)
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
echo %B_YELLOW%WARNING:%RESET% %B_WHITE%%~1%RESET%
echo.
goto :eof

:: Function to show info
:show_info
echo.
echo %B_CYAN%INFO:%RESET% %WHITE%%~1%RESET%
echo.
goto :eof

:: Function to show success
:show_success
echo.
echo %B_GREEN%SUCCESS:%RESET% %WHITE%%~1%RESET%
echo.
goto :eof

:: Check if config.yaml exists
if not exist "%CONFIG_FILE%" (
    echo %B_RED%Error: %CONFIG_FILE% not found in current directory%RESET%
    echo %WHITE%Please run this script from the project root directory.%RESET%
    pause
    exit /b 1
)

:: Backup original config
copy "%CONFIG_FILE%" "%BACKUP_FILE%" >nul

echo.
echo    %P_PURPLE% /\ %RESET%
echo    %P_PURPLE%(  )%RESET%   %B_PURPLE%stackyard Onboarding%RESET% %GRAY%by%RESET% %B_WHITE%diameter-tscd%RESET%
echo   %P_PURPLE% \/ %RESET%
echo %GRAY%----------------------------------------------------------------------%RESET%
echo %B_CYAN%Welcome to the stackyard onboarding setup!%RESET%
echo %GRAY%This script will help you configure your application.%RESET%
echo %GRAY%----------------------------------------------------------------------%RESET%

:: Basic Application Configuration
echo.
echo %B_PURPLE%BASIC APPLICATION CONFIGURATION%RESET%

call :read_input "Enter application name" "My Fancy Go App"
set "APP_NAME=%errorlevel%"

call :read_input "Enter application version" "1.0.0"
set "APP_VERSION=%errorlevel%"

call :read_input "Enter server port" "8080"
set "SERVER_PORT=%errorlevel%"

call :read_input "Enter monitoring port" "9090"
set "MONITORING_PORT=%errorlevel%"

:: Environment Settings
echo.
echo %B_PURPLE%ENVIRONMENT SETTINGS%RESET%

call :read_yes_no "Enable debug mode?" "y"
set "DEBUG_MODE=%errorlevel%"

call :read_yes_no "Enable TUI (Terminal User Interface)?" "y"
set "TUI_MODE=%errorlevel%"

call :read_yes_no "Quiet startup (suppress console logs)?" "n"
set "QUIET_STARTUP=%errorlevel%"

:: Service Configuration
echo.
echo %B_PURPLE%SERVICE CONFIGURATION%RESET%

call :read_yes_no "Enable monitoring dashboard?" "y"
set "ENABLE_MONITORING=%errorlevel%"

call :read_yes_no "Enable API encryption?" "n"
set "ENABLE_ENCRYPTION=%errorlevel%"

:: Infrastructure Configuration
echo.
echo %B_PURPLE%INFRASTRUCTURE CONFIGURATION%RESET%

call :read_yes_no "Enable Redis?" "n"
set "ENABLE_REDIS=%errorlevel%"

echo.
set /p "ENABLE_POSTGRES=Enable PostgreSQL? (single/multi/none) [single]: "
if "%ENABLE_POSTGRES%"=="" set "ENABLE_POSTGRES=single"

call :read_yes_no "Enable Kafka?" "n"
set "ENABLE_KAFKA=%errorlevel%"

call :read_yes_no "Enable MinIO (Object Storage)?" "n"
set "ENABLE_MINIO=%errorlevel%"

:: Apply Configuration
echo.
echo %B_PURPLE%APPLYING CONFIGURATION%RESET%

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
echo %B_PURPLE%SECURITY WARNINGS%RESET%

call :show_warning "Default credentials are configured. You MUST change these before production use:"
echo %B_RED%• PostgreSQL password: 'Mypostgres01'%RESET%
echo %B_RED%• Monitoring password: 'admin'%RESET%
echo %B_RED%• MinIO credentials: 'minioadmin/minioadmin'%RESET%
echo %B_RED%• API secret key: 'super-secret-key'%RESET%

call :show_warning "API obfuscation is enabled. This adds security through obscurity but is not encryption."

if "%ENABLE_ENCRYPTION%"=="true" (
    call :show_warning "Encryption is enabled but no key is set. You need to configure 'encryption.key' in config.yaml"
)

:: Next Steps
echo.
echo %B_PURPLE%NEXT STEPS%RESET%

call :show_info "1. Review and customize config.yaml with your specific settings"
call :show_info "2. Update all default passwords and secrets"
call :show_info "3. Set up your infrastructure (PostgreSQL, Redis, etc.)"
call :show_info "4. Run 'go mod tidy' to ensure dependencies are correct"
call :show_info "5. Build the application using 'scripts\build.bat'"
call :show_info "6. Test the application with 'go run cmd\app\main.go'"

:: Offer to run additional setup
echo.
echo %P_CYAN%Would you like to run additional setup commands?%RESET%
echo %GRAY%This will run 'go mod tidy' and check for build issues.%RESET%
call :read_yes_no "Run setup commands?" "y"
set "RUN_SETUP=%errorlevel%"

if "%RUN_SETUP%"=="true" (
    echo.
    echo %B_PURPLE%RUNNING SETUP COMMANDS%RESET%

    echo.
    echo %P_CYAN%Running 'go mod tidy'...%RESET%
    go mod tidy
    if %ERRORLEVEL% EQU 0 (
        call :show_success "Dependencies updated successfully"
    ) else (
        call :show_warning "Failed to update dependencies - you may need to check your Go installation"
    )

    echo.
    echo %P_CYAN%Checking build...%RESET%
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
echo %GRAY%======================================================================%RESET%
echo  %B_PURPLE%ONBOARDING COMPLETE!%RESET% %P_GREEN%Your app is ready to go!%RESET%
echo %GRAY%======================================================================%RESET%
echo.
echo %B_CYAN%Backup created:%RESET% %B_WHITE%%BACKUP_FILE%%RESET%
echo %B_CYAN%Configuration:%RESET% %B_WHITE%%CONFIG_FILE%%RESET%
echo.
echo %B_GREEN%Happy coding!%RESET%
echo.

pause
endlocal
