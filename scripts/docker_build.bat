@echo off
cls

setlocal EnableDelayedExpansion
set "DEFAULT_APP_NAME=stackyard"
set "DEFAULT_IMAGE_NAME=myapp"
set "DEFAULT_TARGET=all"

REM Configuration from parameters or defaults
if "%~1"=="" (
    set "APP_NAME=%DEFAULT_APP_NAME%"
) else (
    set "APP_NAME=%~1"
)

if "%~2"=="" (
    set "IMAGE_NAME=%DEFAULT_IMAGE_NAME%"
) else (
    set "IMAGE_NAME=%~2"
)

if "%~3"=="" (
    set "TARGET=%DEFAULT_TARGET%"
) else (
    set "TARGET=%~3"
)

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
set "P_GREEN=%ESC%[38;5;108m"
set "B_GREEN=%ESC%[1;38;5;108m"
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
echo    %P_PURPLE% /\ %RESET%
echo    %P_PURPLE%(  )%RESET%   %B_PURPLE%Docker Builder%RESET% %GRAY%by%RESET% %B_WHITE%diameter-tscd%RESET%
echo   %P_PURPLE% \/ %RESET%
echo %GRAY%----------------------------------------------------------------------%RESET%
echo    %B_CYAN%App Name:%RESET% %B_WHITE%%APP_NAME%%RESET%
echo    %B_CYAN%Image Name:%RESET% %B_WHITE%%IMAGE_NAME%%RESET%
echo    %B_CYAN%Target:%RESET% %B_WHITE%%TARGET%%RESET%
echo %GRAY%----------------------------------------------------------------------%RESET%

REM Check if Dockerfile exists
if not exist "Dockerfile" (
    echo    %B_RED%x Dockerfile not found in current directory%RESET%
    exit /b 1
)

REM Check if docker is available
docker --version >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo    %B_RED%x Docker is not installed or not in PATH%RESET%
    exit /b 1
)

REM Validate target
if "%TARGET%"=="all" goto :valid_target
if "%TARGET%"=="test" goto :valid_target
if "%TARGET%"=="dev" goto :valid_target
if "%TARGET%"=="prod" goto :valid_target
if "%TARGET%"=="prod-slim" goto :valid_target
if "%TARGET%"=="prod-minimal" goto :valid_target
if "%TARGET%"=="ultra-prod" goto :valid_target
if "%TARGET%"=="ultra-all" goto :valid_target
if "%TARGET%"=="ultra-dev" goto :valid_target
if "%TARGET%"=="ultra-test" goto :valid_target
echo    %B_RED%x Invalid target: %TARGET%%RESET%
echo    %B_CYAN%Valid targets: all, test, dev, prod, prod-slim, prod-minimal, ultra-prod, ultra-all, ultra-dev, ultra-test%RESET%
exit /b 1

:valid_target

REM Initialize step counter
set STEP=1

REM Calculate total steps
if "%TARGET%"=="all" (
    set TOTAL_STEPS=4
)
if "%TARGET%"=="test" (
    set TOTAL_STEPS=2
)
if "%TARGET%"=="dev" (
    set TOTAL_STEPS=1
)
if "%TARGET%"=="prod" (
    set TOTAL_STEPS=1
)
if "%TARGET%"=="ultra-prod" (
    set TOTAL_STEPS=1
)
if "%TARGET%"=="ultra-all" (
    set TOTAL_STEPS=4
)
if "%TARGET%"=="ultra-dev" (
    set TOTAL_STEPS=1
)
if "%TARGET%"=="ultra-test" (
    set TOTAL_STEPS=2
)

REM Build Test Stage (always needed for test target or all)
if "%TARGET%"=="test" goto :build_test
if "%TARGET%"=="all" goto :build_test
if "%TARGET%"=="ultra-test" goto :build_test
if "%TARGET%"=="ultra-all" goto :build_test
goto :skip_test

:build_test
echo %B_PURPLE%[%STEP%/%TOTAL_STEPS%]%RESET% %P_CYAN%Building test image...%RESET%
docker build --target test -t "%IMAGE_NAME%:test" .
if %ERRORLEVEL% equ 0 (
    echo    %B_GREEN%+ Test image built:%RESET% %B_WHITE%%IMAGE_NAME%:test%RESET%
) else (
    echo    %B_RED%x Test build failed%RESET%
    exit /b 1
)
set /a STEP+=1

REM Run Tests (only for test target or all)
if "%TARGET%"=="test" goto :run_tests
if "%TARGET%"=="all" goto :run_tests
if "%TARGET%"=="ultra-test" goto :run_tests
if "%TARGET%"=="ultra-all" goto :run_tests
goto :skip_tests

:run_tests
echo %B_PURPLE%[%STEP%/%TOTAL_STEPS%]%RESET% %P_CYAN%Running tests...%RESET%
docker run --rm "%IMAGE_NAME%:test"
if %ERRORLEVEL% equ 0 (
    echo    %B_GREEN%+ Tests passed%RESET%
) else (
    echo    %B_RED%x Tests failed%RESET%
    exit /b 1
)
set /a STEP+=1

:skip_tests
:skip_test

REM Build Development Stage
if "%TARGET%"=="dev" goto :build_dev
if "%TARGET%"=="all" goto :build_dev
goto :skip_dev

:build_dev
echo %B_PURPLE%[%STEP%/%TOTAL_STEPS%]%RESET% %P_CYAN%Building development image...%RESET%
docker build --target dev -t "%IMAGE_NAME%:dev" .
if %ERRORLEVEL% equ 0 (
    echo    %B_GREEN%+ Development image built:%RESET% %B_WHITE%%IMAGE_NAME%:dev%RESET%
) else (
    echo    %B_RED%x Development build failed%RESET%
    exit /b 1
)
set /a STEP+=1

:skip_dev

REM Build Ultra Development Stage
if "%TARGET%"=="ultra-dev" goto :build_ultra_dev
if "%TARGET%"=="ultra-all" goto :build_ultra_dev
goto :skip_ultra_dev

:build_ultra_dev
echo %B_PURPLE%[%STEP%/%TOTAL_STEPS%]%RESET% %P_CYAN%Building ultra development image...%RESET%
docker build --target ultra-dev -t "%IMAGE_NAME%:dev" .
if %ERRORLEVEL% equ 0 (
    echo    %B_GREEN%+ Ultra development image built:%RESET% %B_WHITE%%IMAGE_NAME%:dev%RESET%
) else (
    echo    %B_RED%x Ultra development build failed%RESET%
    exit /b 1
)
set /a STEP+=1

:skip_ultra_dev

REM Build Production Stage
if "%TARGET%"=="prod" goto :build_prod
if "%TARGET%"=="all" goto :build_prod
goto :skip_prod

:build_prod
echo %B_PURPLE%[%STEP%/%TOTAL_STEPS%]%RESET% %P_CYAN%Building production image...%RESET%
docker build --target prod -t "%IMAGE_NAME%:latest" .
if %ERRORLEVEL% equ 0 (
    echo    %B_GREEN%+ Production image built:%RESET% %B_WHITE%%IMAGE_NAME%:latest%RESET%
) else (
    echo    %B_RED%x Production build failed%RESET%
    exit /b 1
)
set /a STEP+=1

:skip_prod

REM Build Slim Production Stage
if "%TARGET%"=="prod-slim" goto :build_prod_slim
goto :skip_prod_slim

:build_prod_slim
echo %B_PURPLE%[%STEP%/%TOTAL_STEPS%]%RESET% %P_CYAN%Building slim production image...%RESET%
docker build --target prod-slim -t "%IMAGE_NAME%:slim" .
if %ERRORLEVEL% equ 0 (
    echo    %B_GREEN%+ Slim production image built:%RESET% %B_WHITE%%IMAGE_NAME%:slim%RESET%
) else (
    echo    %B_RED%x Slim production build failed%RESET%
    exit /b 1
)
set /a STEP+=1

:skip_prod_slim

REM Build Minimal Production Stage
if "%TARGET%"=="prod-minimal" goto :build_prod_minimal
goto :skip_prod_minimal

:build_prod_minimal
echo %B_PURPLE%[%STEP%/%TOTAL_STEPS%]%RESET% %P_CYAN%Building minimal production image...%RESET%
docker build --target prod-minimal -t "%IMAGE_NAME%:minimal" .
if %ERRORLEVEL% equ 0 (
    echo    %B_GREEN%+ Minimal production image built:%RESET% %B_WHITE%%IMAGE_NAME%:minimal%RESET%
) else (
    echo    %B_RED%x Minimal production build failed%RESET%
    exit /b 1
)
set /a STEP+=1

:skip_prod_minimal

REM Build Ultra Production Stage (for ultra-all and ultra-prod)
if "%TARGET%"=="ultra-all" goto :build_ultra_prod
if "%TARGET%"=="ultra-prod" goto :build_ultra_prod
goto :skip_ultra_prod

:build_ultra_prod
echo %B_PURPLE%[%STEP%/%TOTAL_STEPS%]%RESET% %P_CYAN%Building ultra production image...%RESET%
docker build --target ultra-prod -t "%IMAGE_NAME%:ultra" .
if %ERRORLEVEL% equ 0 (
    echo    %B_GREEN%+ Ultra production image built:%RESET% %B_WHITE%%IMAGE_NAME%:ultra%RESET%
) else (
    echo    %B_RED%x Ultra production build failed%RESET%
    exit /b 1
)
set /a STEP+=1

:skip_ultra_prod

REM Optional: Clean up intermediate images
echo %B_PURPLE%[Bonus]%RESET% %P_CYAN%Cleaning up dangling images...%RESET%
docker image prune -f >nul 2>&1
if %ERRORLEVEL% equ 0 (
    echo    %B_GREEN%+ Cleanup completed%RESET%
) else (
    echo    %GRAY%- Cleanup skipped%RESET%
)

echo.
echo %GRAY%======================================================================%RESET%
echo  %B_PURPLE%SUCCESS!%RESET% %P_GREEN%Docker images ready:%RESET%

REM Show only the images that were actually built
if "%TARGET%"=="test" echo    %B_WHITE%%IMAGE_NAME%:test%RESET%     %GRAY%(testing)%RESET%
if "%TARGET%"=="all" echo    %B_WHITE%%IMAGE_NAME%:test%RESET%     %GRAY%(testing)%RESET%
if "%TARGET%"=="ultra-test" echo    %B_WHITE%%IMAGE_NAME%:test%RESET%     %GRAY%(testing)%RESET%
if "%TARGET%"=="ultra-all" echo    %B_WHITE%%IMAGE_NAME%:test%RESET%     %GRAY%(testing)%RESET%
if "%TARGET%"=="dev" echo    %B_WHITE%%IMAGE_NAME%:dev%RESET%      %GRAY%(development)%RESET%
if "%TARGET%"=="all" echo    %B_WHITE%%IMAGE_NAME%:dev%RESET%      %GRAY%(development)%RESET%
if "%TARGET%"=="ultra-dev" echo    %B_WHITE%%IMAGE_NAME%:dev%RESET%      %GRAY%(development)%RESET%
if "%TARGET%"=="ultra-all" echo    %B_WHITE%%IMAGE_NAME%:dev%RESET%      %GRAY%(development)%RESET%
if "%TARGET%"=="prod" echo    %B_WHITE%%IMAGE_NAME%:latest%RESET%  %GRAY%(production)%RESET%
if "%TARGET%"=="all" echo    %B_WHITE%%IMAGE_NAME%:latest%RESET%  %GRAY%(production)%RESET%
if "%TARGET%"=="prod-slim" echo    %B_WHITE%%IMAGE_NAME%:slim%RESET%    %GRAY%(slim-production)%RESET%
if "%TARGET%"=="prod-minimal" echo    %B_WHITE%%IMAGE_NAME%:minimal%RESET% %GRAY%(minimal-production)%RESET%
if "%TARGET%"=="ultra-prod" echo    %B_WHITE%%IMAGE_NAME%:ultra%RESET%    %GRAY%(ultra-production)%RESET%
if "%TARGET%"=="ultra-all" echo    %B_WHITE%%IMAGE_NAME%:ultra%RESET%    %GRAY%(ultra-production)%RESET%

echo %GRAY%======================================================================%RESET%
echo.
echo %B_CYAN%Usage examples:%RESET%

REM Show relevant usage examples based on what was built
if "%TARGET%"=="dev" echo   %GRAY%# Run development container%RESET%
if "%TARGET%"=="dev" echo   %B_WHITE%docker run -p 8080:8080 -p 9090:9090 %IMAGE_NAME%:dev%RESET%
if "%TARGET%"=="dev" echo.
if "%TARGET%"=="all" echo   %GRAY%# Run development container%RESET%
if "%TARGET%"=="all" echo   %B_WHITE%docker run -p 8080:8080 -p 9090:9090 %IMAGE_NAME%:dev%RESET%
if "%TARGET%"=="all" echo.
if "%TARGET%"=="ultra-dev" echo   %GRAY%# Run development container%RESET%
if "%TARGET%"=="ultra-dev" echo   %B_WHITE%docker run -p 8080:8080 -p 9090:9090 %IMAGE_NAME%:dev%RESET%
if "%TARGET%"=="ultra-dev" echo.
if "%TARGET%"=="ultra-all" echo   %GRAY%# Run development container%RESET%
if "%TARGET%"=="ultra-all" echo   %B_WHITE%docker run -p 8080:8080 -p 9090:9090 %IMAGE_NAME%:dev%RESET%
if "%TARGET%"=="ultra-all" echo.
if "%TARGET%"=="prod" echo   %GRAY%# Run production container%RESET%
if "%TARGET%"=="prod" echo   %B_WHITE%docker run -p 8080:8080 -p 9090:9090 %IMAGE_NAME%:latest%RESET%
if "%TARGET%"=="prod" echo.
if "%TARGET%"=="all" echo   %GRAY%# Run production container%RESET%
if "%TARGET%"=="all" echo   %B_WHITE%docker run -p 8080:8080 -p 9090:9090 %IMAGE_NAME%:latest%RESET%
if "%TARGET%"=="all" echo.
if "%TARGET%"=="ultra-prod" echo   %GRAY%# Run ultra production container%RESET%
if "%TARGET%"=="ultra-prod" echo   %B_WHITE%docker run -p 8080:8080 -p 9090:9090 %IMAGE_NAME%:ultra%RESET%
if "%TARGET%"=="ultra-prod" echo.
if "%TARGET%"=="ultra-all" echo   %GRAY%# Run ultra production container%RESET%
if "%TARGET%"=="ultra-all" echo   %B_WHITE%docker run -p 8080:8080 -p 9090:9090 %IMAGE_NAME%:ultra%RESET%
if "%TARGET%"=="ultra-all" echo.
if "%TARGET%"=="test" echo   %GRAY%# Run tests%RESET%
if "%TARGET%"=="test" echo   %B_WHITE%docker run --rm %IMAGE_NAME%:test%RESET%
if "%TARGET%"=="all" echo   %GRAY%# Run tests%RESET%
if "%TARGET%"=="all" echo   %B_WHITE%docker run --rm %IMAGE_NAME%:test%RESET%
if "%TARGET%"=="ultra-test" echo   %GRAY%# Run tests%RESET%
if "%TARGET%"=="ultra-test" echo   %B_WHITE%docker run --rm %IMAGE_NAME%:test%RESET%
if "%TARGET%"=="ultra-all" echo   %GRAY%# Run tests%RESET%
if "%TARGET%"=="ultra-all" echo   %B_WHITE%docker run --rm %IMAGE_NAME%:test%RESET%
endlocal
