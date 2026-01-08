#!/bin/bash

# Clear the terminal screen
clear

# Configuration
DIST_DIR="dist"
APP_NAME="stackyard"
MAIN_PATH="./cmd/app/main.go"

# Define ANSI Colors
RESET="\033[0m"
BOLD="\033[1m"
DIM="\033[2m"
UNDERLINE="\033[4m"

# Fancy Pastel Palette (main color: #8daea5)
P_PURPLE="\033[38;5;108m"
B_PURPLE="\033[1;38;5;108m"
P_CYAN="\033[38;5;117m"
B_CYAN="\033[1;38;5;117m"
P_GREEN="\033[38;5;108m"
B_GREEN="\033[1;38;5;108m"
P_YELLOW="\033[93m"
B_YELLOW="\033[1;93m"
P_RED="\033[91m"
B_RED="\033[1;91m"
GRAY="\033[38;5;242m"
WHITE="\033[97m"
B_WHITE="\033[1;97m"

# Robustly switch to project root (one level up from this script)
cd "$(dirname "$0")/.." || exit 1

echo ""
echo -e "   ${P_PURPLE} /\ ${RESET}"
echo -e "   ${P_PURPLE}(  )${RESET}   ${B_PURPLE}${APP_NAME} Builder${RESET} ${GRAY}by${RESET} ${B_WHITE}diameter-tscd${RESET}"
echo -e "  ${P_PURPLE} \/ ${RESET}"
echo -e "${GRAY}----------------------------------------------------------------------${RESET}"

# 0. Check required tools
echo -e "${B_PURPLE}[0/6]${RESET} ${P_CYAN}Checking required tools...${RESET}"

# Check goversioninfo
if ! command -v goversioninfo &> /dev/null; then
    echo -e "   ${B_YELLOW}! goversioninfo not found. Skipping version info generation.${RESET}"
    USE_GOVERSIONINFO=false
else
    echo -e "   ${B_GREEN}+ goversioninfo found${RESET}"
    USE_GOVERSIONINFO=true
fi

# Check garble
if ! command -v garble &> /dev/null; then
    echo -e "   ${B_YELLOW}! garble not found. Installing...${RESET}"
    go install mvdan.cc/garble@latest
    if [ $? -ne 0 ]; then
        echo -e "   ${B_RED}x Failed to install garble${RESET}"
        exit 1
    fi
    echo -e "   ${B_GREEN}+ garble installed${RESET}"
else
    echo -e "   ${B_GREEN}+ garble found${RESET}"
fi

# Ask user about garble build
echo -e "${B_YELLOW}Use garble build for obfuscation? (y/N, timeout 10s): ${RESET}"
read -t 10 -n 1 -r choice
echo ""
if [[ $choice =~ ^[Yy]$ ]]; then
    USE_GARBLE=true
    echo -e "${B_GREEN}+ Using garble build${RESET}"
else
    USE_GARBLE=false
    echo -e "${B_CYAN}+ Using regular go build${RESET}"
fi

# 1. Generate Timestamp
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_ROOT="${DIST_DIR}/backups"
BACKUP_PATH="${BACKUP_ROOT}/${TIMESTAMP}"

# 2. Stop running process
echo -e "${B_PURPLE}[1/6]${RESET} ${P_CYAN}Checking for running process...${RESET}"
if pgrep -x "$APP_NAME" >/dev/null; then
    echo -e "   ${B_YELLOW}! App is running. Stopping...${RESET}"
    pkill -x "$APP_NAME"
    sleep 1
else
    echo -e "   ${B_GREEN}+ App is not running.${RESET}"
fi

# 3. Backup Old Files
echo -e "${B_PURPLE}[2/6]${RESET} ${P_CYAN}Backing up old files...${RESET}"
if [ -d "$DIST_DIR" ]; then
    mkdir -p "$BACKUP_PATH"

    # Move old binary (check for both plain and .exe just in case)
    if [ -f "$DIST_DIR/$APP_NAME" ]; then
        echo -e "   ${GRAY}- Moving old binary...${RESET}"
        mv "$DIST_DIR/$APP_NAME" "$BACKUP_PATH/"
    elif [ -f "$DIST_DIR/$APP_NAME.exe" ]; then
        echo -e "   ${GRAY}- Moving old binary (.exe)...${RESET}"
        mv "$DIST_DIR/$APP_NAME.exe" "$BACKUP_PATH/"
    fi

    if [ -f "$DIST_DIR/config.yaml" ]; then
        mv "$DIST_DIR/config.yaml" "$BACKUP_PATH/"
    fi
    if [ -f "$DIST_DIR/banner.txt" ]; then
        mv "$DIST_DIR/banner.txt" "$BACKUP_PATH/"
    fi
    if [ -f "$DIST_DIR/monitoring_users.db" ]; then
        echo -e "   ${GRAY}- Backing up database...${RESET}"
        mv "$DIST_DIR/monitoring_users.db" "$BACKUP_PATH/"
    fi
    if [ -d "$DIST_DIR/web" ]; then
        echo -e "   ${GRAY}- Moving old web assets...${RESET}"
        mv "$DIST_DIR/web" "$BACKUP_PATH/"
    fi
    
    echo -e "   ${B_GREEN}+ Backup created at:${RESET} ${B_WHITE}${BACKUP_PATH}${RESET}"
else
    echo -e "   ${GRAY}+ No existing dist directory. Skipping backup.${RESET}"
    mkdir -p "$DIST_DIR"
fi

# 6. Archive Backup
echo -e "${B_PURPLE}[3/6]${RESET} ${P_CYAN}Archiving backup...${RESET}"
if [ -d "$BACKUP_PATH" ]; then
    cd "$BACKUP_ROOT" || exit 1
    zip -r "${TIMESTAMP}.zip" "$TIMESTAMP"
    rm -rf "$TIMESTAMP"
    cd - >/dev/null || exit 1  # Return to previous directory
    echo -e "   ${B_GREEN}+ Backup archived:${RESET} ${B_WHITE}${BACKUP_ROOT}/${TIMESTAMP}.zip${RESET}"
else
    echo -e "   ${GRAY}+ No backup created. Skipping archive.${RESET}"
fi

# Ensure dist directory
mkdir -p "$DIST_DIR"

# 4. Build
echo -e "${B_PURPLE}[4/6]${RESET} ${P_CYAN}Building Go binary...${RESET}"
if [ "$USE_GOVERSIONINFO" = true ]; then
    goversioninfo -platform-specific
else
    echo -e "   ${GRAY}+ Skipping goversioninfo (not available)${RESET}"
fi
if [ "$USE_GARBLE" = true ]; then
    garble build -ldflags="-s -w" -o "$DIST_DIR/$APP_NAME" "$MAIN_PATH"
else
    go build -ldflags="-s -w" -o "$DIST_DIR/$APP_NAME" "$MAIN_PATH"
fi
if [ $? -ne 0 ]; then
    echo -e "   ${B_RED}x Build FAILED! Exit code: $?${RESET}"
    exit $?
fi
echo -e "   ${B_GREEN}+ Build successful:${RESET} ${B_WHITE}${DIST_DIR}/${APP_NAME}${RESET}"

# 5. Copy Assets
echo -e "${B_PURPLE}[5/6]${RESET} ${P_CYAN}Copying assets...${RESET}"

if [ -d "web" ]; then
    echo -e "   ${B_GREEN}+ Copying web folder...${RESET}"
    cp -r "web" "$DIST_DIR/web"
fi

if [ -f "config.yaml" ]; then
    echo -e "   ${B_GREEN}+ Copying config.yaml...${RESET}"
    cp "config.yaml" "$DIST_DIR/"
fi

if [ -f "banner.txt" ]; then
    echo -e "   ${B_GREEN}+ Copying banner.txt...${RESET}"
    cp "banner.txt" "$DIST_DIR/"
fi

if [ -f "monitoring_users.db" ]; then
    echo -e "   ${B_GREEN}+ Copying monitoring_users.db...${RESET}"
    cp "monitoring_users.db" "$DIST_DIR/"
fi

echo ""
echo -e "${GRAY}======================================================================${RESET}"
echo -e " ${B_PURPLE}SUCCESS!${RESET} ${P_GREEN}Build ready at:${RESET} ${UNDERLINE}${B_WHITE}${DIST_DIR}/${RESET}"
echo -e "${GRAY}======================================================================${RESET}"
