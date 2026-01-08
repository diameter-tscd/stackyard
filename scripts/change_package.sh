#!/bin/bash

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

# Script to change the Go module package name
# Usage: ./change_package.sh <new_module_name>

if [ $# -ne 1 ]; then
  echo "Usage: $0 <new_module_name>"
  echo "Example: $0 github.com/user/new-project"
  exit 1
fi

NEW_MODULE=$1

# Get current module name from go.mod
CURRENT_MODULE=$(grep '^module ' go.mod | awk '{print $2}')

if [ -z "$CURRENT_MODULE" ]; then
  echo "Error: Could not find module declaration in go.mod"
  exit 1
fi

echo "Changing module from '$CURRENT_MODULE' to '$NEW_MODULE'"

# Change module name in go.mod
sed -i.bak "s|^module $CURRENT_MODULE|module $NEW_MODULE|" go.mod

if [ $? -ne 0 ]; then
  echo "Error: Failed to update go.mod"
  exit 1
fi

# Update all import paths in .go files
find . -name "*.go" -type f -exec sed -i.bak "s|$CURRENT_MODULE|$NEW_MODULE|g" {} +

echo "Successfully changed module name and updated imports"
echo "Note: Backup files (*.bak) have been created. You can remove them if everything looks good."
