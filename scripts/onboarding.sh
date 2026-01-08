#!/bin/bash

# Onboarding Script for stackyard
# This script helps set up the application for the first time

# Clear the terminal screen
clear

# Check if build script exists
if [ ! -f "./scripts/build.sh" ]; then
    echo -e "${B_RED}Error: build.sh not found in scripts/ directory${RESET}"
    exit 1
fi

# Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
    echo -e "${B_RED}Error: Go is not installed or not in PATH${RESET}"
    echo -e "${WHITE}Please install Go from https://golang.org/dl/${RESET}"
    exit 1
fi

# Check Go version (minimum 1.24)
go_version=$(go version | awk '{print $3}' | sed 's/go//')
if [ "$(printf '%s\n' "$go_version" "1.24" | sort -V | head -n1)" = "$go_version" ]; then
    echo -e "${B_RED}Error: Go version $go_version is too old. Minimum required is 1.24${RESET}"
    echo -e "${WHITE}Please upgrade Go from https://golang.org/dl/${RESET}"
    exit 1
fi

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
P_GREEN="\033[38;5;46m"
B_GREEN="\033[1;38;5;46m"
P_YELLOW="\033[93m"
B_YELLOW="\033[1;93m"
P_RED="\033[91m"
B_RED="\033[1;91m"
GRAY="\033[38;5;242m"
WHITE="\033[97m"
B_WHITE="\033[1;97m"

# Function to read user input with default value
read_input() {
    local prompt="$1"
    local default="$2"
    local input

    echo -e "${P_CYAN}$prompt${RESET}"
    if [ -n "$default" ]; then
        echo -e "${GRAY}Default: $default${RESET}"
    fi
    echo -ne "${B_WHITE}> ${RESET}"
    read input

    if [ -z "$input" ] && [ -n "$default" ]; then
        input="$default"
    fi

    echo "$input"
}

# Function to read yes/no with default
read_yes_no() {
    local prompt="$1"
    local default="$2"
    local input

    echo -e "${P_CYAN}$prompt${RESET}"
    if [ "$default" = "y" ]; then
        echo -e "${GRAY}Default: Yes${RESET}"
    else
        echo -e "${GRAY}Default: No${RESET}"
    fi
    echo -ne "${B_WHITE}(y/n) > ${RESET}"
    read input

    if [ -z "$input" ]; then
        input="$default"
    fi

    case "$input" in
        y|Y|yes|Yes|YES) echo "true" ;;
        *) echo "false" ;;
    esac
}

# Function to update config value
update_config() {
    local key="$1"
    local value="$2"

    # Escape special characters for sed
    value=$(printf '%s\n' "$value" | sed 's/[[\.*^$(){}?+|/]/\\&/g')

    if grep -q "^$key:" config.yaml; then
        # Update existing key
        sed -i.bak "s|^$key:.*|$key: \"$value\"|" config.yaml
    else
        # Add new key (this is a simple implementation - might need adjustment for complex YAML)
        echo "Warning: Could not find $key in config.yaml - you may need to add it manually"
    fi
}

# Function to update boolean config value
update_config_bool() {
    local key="$1"
    local value="$2"

    if grep -q "^$key:" config.yaml; then
        # Update existing key
        sed -i.bak "s|^$key:.*|$key: $value|" config.yaml
    else
        echo "Warning: Could not find $key in config.yaml - you may need to add it manually"
    fi
}

# Function to show warning
show_warning() {
    local message="$1"
    echo -e "${B_YELLOW}WARNING:${RESET} ${B_WHITE}$message${RESET}"
    echo ""
}

# Function to show info
show_info() {
    local message="$1"
    echo -e "${B_CYAN}INFO:${RESET} ${WHITE}$message${RESET}"
    echo ""
}

# Function to show success
show_success() {
    local message="$1"
    echo -e "${B_GREEN}SUCCESS:${RESET} ${WHITE}$message${RESET}"
    echo ""
}

# Check if config.yaml exists
if [ ! -f "config.yaml" ]; then
    echo -e "${B_RED}Error: config.yaml not found in current directory${RESET}"
    echo -e "${WHITE}Please run this script from the project root directory.${RESET}"
    exit 1
fi

# Backup original config
cp config.yaml config.yaml.backup

echo ""
echo -e "   ${P_PURPLE}(\_/)${RESET}"
echo -e "   ${P_PURPLE}(o.o)${RESET}   ${B_PURPLE}stackyard Onboarding${RESET} ${GRAY}by${RESET} ${B_WHITE}diameter-tscd${RESET}"
echo -e "  ${P_PURPLE}c(\")(\")${RESET}"
echo -e "${GRAY}----------------------------------------------------------------------${RESET}"
echo -e "${B_CYAN}Welcome to the stackyard onboarding setup!${RESET}"
echo -e "${GRAY}This script will help you configure your application.${RESET}"
echo -e "${GRAY}----------------------------------------------------------------------${RESET}"
echo ""

# Basic Application Configuration
echo -e "${B_PURPLE}BASIC APPLICATION CONFIGURATION${RESET}"
echo ""

APP_NAME=$(read_input "Enter application name" "My Fancy Go App")
APP_VERSION=$(read_input "Enter application version" "1.0.0")
SERVER_PORT=$(read_input "Enter server port" "8080")
MONITORING_PORT=$(read_input "Enter monitoring port" "9090")

echo ""

# Environment Settings
echo -e "${B_PURPLE}ENVIRONMENT SETTINGS${RESET}"
echo ""

DEBUG_MODE=$(read_yes_no "Enable debug mode?" "y")
TUI_MODE=$(read_yes_no "Enable TUI (Terminal User Interface)?" "y")
QUIET_STARTUP=$(read_yes_no "Quiet startup (suppress console logs)?" "n")

echo ""

# Service Configuration
echo -e "${B_PURPLE}SERVICE CONFIGURATION${RESET}"
echo ""

ENABLE_MONITORING=$(read_yes_no "Enable monitoring dashboard?" "y")
ENABLE_ENCRYPTION=$(read_yes_no "Enable API encryption?" "n")

echo ""

# Infrastructure Configuration
echo -e "${B_PURPLE}INFRASTRUCTURE CONFIGURATION${RESET}"
echo ""

ENABLE_REDIS=$(read_yes_no "Enable Redis?" "n")
ENABLE_POSTGRES=$(read_input "Enable PostgreSQL? (single/multi/none)" "single")
ENABLE_KAFKA=$(read_yes_no "Enable Kafka?" "n")
ENABLE_MINIO=$(read_yes_no "Enable MinIO (Object Storage)?" "n")

echo ""

# Apply Configuration
echo -e "${B_PURPLE}APPLYING CONFIGURATION${RESET}"
echo ""

# Update basic config
update_config "app.name" "$APP_NAME"
update_config "app.version" "$APP_VERSION"
update_config "server.port" "$SERVER_PORT"
update_config "monitoring.port" "$MONITORING_PORT"

# Update boolean configs
update_config_bool "app.debug" "$DEBUG_MODE"
update_config_bool "app.enable_tui" "$TUI_MODE"
update_config_bool "app.quiet_startup" "$QUIET_STARTUP"
update_config_bool "monitoring.enabled" "$ENABLE_MONITORING"
update_config_bool "encryption.enabled" "$ENABLE_ENCRYPTION"
update_config_bool "redis.enabled" "$ENABLE_REDIS"
update_config_bool "kafka.enabled" "$ENABLE_KAFKA"

# Handle PostgreSQL configuration
if [ "$ENABLE_POSTGRES" = "single" ]; then
    update_config_bool "postgres.enabled" "true"
    # Note: Single connection config would need manual setup
elif [ "$ENABLE_POSTGRES" = "multi" ]; then
    update_config_bool "postgres.enabled" "true"
    # Multi-connection is already configured in the template
else
    update_config_bool "postgres.enabled" "false"
fi

# Handle MinIO
if [ "$ENABLE_MINIO" = "true" ]; then
    update_config_bool "monitoring.minio.enabled" "true"
else
    update_config_bool "monitoring.minio.enabled" "false"
fi

show_success "Configuration updated successfully!"

# Security Warnings
echo -e "${B_PURPLE}SECURITY WARNINGS${RESET}"
echo ""

show_warning "Default credentials are configured. You MUST change these before production use:"
echo -e "${B_RED}• PostgreSQL password: 'Mypostgres01'${RESET}"
echo -e "${B_RED}• Monitoring password: 'admin'${RESET}"
echo -e "${B_RED}• MinIO credentials: 'minioadmin/minioadmin'${RESET}"
echo -e "${B_RED}• API secret key: 'super-secret-key'${RESET}"
echo ""

show_warning "API obfuscation is enabled. This adds security through obscurity but is not encryption."
echo ""

if [ "$ENABLE_ENCRYPTION" = "true" ]; then
    show_warning "Encryption is enabled but no key is set. You need to configure 'encryption.key' in config.yaml"
    echo ""
fi

# Next Steps
echo -e "${B_PURPLE}NEXT STEPS${RESET}"
echo ""

show_info "1. Review and customize config.yaml with your specific settings"
show_info "2. Update all default passwords and secrets"
show_info "3. Set up your infrastructure (PostgreSQL, Redis, etc.)"
show_info "4. Run 'go mod tidy' to ensure dependencies are correct"
show_info "5. Build the application using './scripts/build.sh'"
show_info "6. Test the application with 'go run cmd/app/main.go'"

echo ""

# Offer to run additional setup
echo -e "${P_CYAN}Would you like to run additional setup commands?${RESET}"
echo -e "${GRAY}This will run 'go mod tidy' and check for build issues.${RESET}"
RUN_SETUP=$(read_yes_no "Run setup commands?" "y")

if [ "$RUN_SETUP" = "true" ]; then
    echo ""
    echo -e "${B_PURPLE}RUNNING SETUP COMMANDS${RESET}"
    echo ""

    echo -e "${P_CYAN}Running 'go mod tidy'...${RESET}"
    if go mod tidy; then
        show_success "Dependencies updated successfully"
    else
        show_warning "Failed to update dependencies - you may need to check your Go installation"
    fi

    echo -e "${P_CYAN}Checking build...${RESET}"
    if go build -o /tmp/test-build ./cmd/app/main.go; then
        show_success "Build test successful"
        rm /tmp/test-build 2>/dev/null
    else
        show_warning "Build failed - check your configuration and dependencies"
    fi

    echo ""
fi

# Final message
echo -e "${GRAY}======================================================================${RESET}"
echo -e " ${B_PURPLE}ONBOARDING COMPLETE!${RESET} ${P_GREEN}Your app is ready to go!${RESET}"
echo -e "${GRAY}======================================================================${RESET}"
echo ""
echo -e "${B_CYAN}Backup created:${RESET} ${B_WHITE}config.yaml.backup${RESET}"
echo -e "${B_CYAN}Configuration:${RESET} ${B_WHITE}config.yaml${RESET}"
echo ""
echo -e "${B_GREEN}Happy coding!${RESET}"
echo ""

# Restore backup on error (if something went wrong)
trap 'echo ""; echo -e "${B_RED}An error occurred. Restoring backup...${RESET}"; cp config.yaml.backup config.yaml; exit 1' ERR
