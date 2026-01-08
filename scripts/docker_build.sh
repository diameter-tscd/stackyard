#!/bin/bash

# Clear the terminal screen
clear

# Default configuration
DEFAULT_APP_NAME="stackyard"
DEFAULT_IMAGE_NAME="myapp"
DEFAULT_TARGET="all"

# Configuration from parameters or defaults
APP_NAME="${1:-$DEFAULT_APP_NAME}"
IMAGE_NAME="${2:-$DEFAULT_IMAGE_NAME}"
TARGET="${3:-$DEFAULT_TARGET}"

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
echo -e "   ${P_PURPLE}(  )${RESET}   ${B_PURPLE}Docker Builder${RESET} ${GRAY}by${RESET} ${B_WHITE}diameter-tscd${RESET}"
echo -e "  ${P_PURPLE} \/ ${RESET}"
echo -e "${GRAY}----------------------------------------------------------------------${RESET}"
echo -e "   ${B_CYAN}App Name:${RESET} ${B_WHITE}${APP_NAME}${RESET}"
echo -e "   ${B_CYAN}Image Name:${RESET} ${B_WHITE}${IMAGE_NAME}${RESET}"
echo -e "   ${B_CYAN}Target:${RESET} ${B_WHITE}${TARGET}${RESET}"
echo -e "${GRAY}----------------------------------------------------------------------${RESET}"

# Check if Dockerfile exists
if [ ! -f "Dockerfile" ]; then
    echo -e "   ${B_RED}x Dockerfile not found in current directory${RESET}"
    exit 1
fi

# Check if docker is available
if ! command -v docker &> /dev/null; then
    echo -e "   ${B_RED}x Docker is not installed or not in PATH${RESET}"
    exit 1
fi

# Validate target
case "$TARGET" in
    "all"|"test"|"dev"|"prod"|"prod-slim"|"prod-minimal"|"ultra-prod"|"ultra-all"|"ultra-dev"|"ultra-test")
        ;;
    *)
        echo -e "   ${B_RED}x Invalid target: ${TARGET}${RESET}"
        echo -e "   ${B_CYAN}Valid targets: all, test, dev, prod, prod-slim, prod-minimal, ultra-prod, ultra-all, ultra-dev, ultra-test${RESET}"
        exit 1
        ;;
esac

STEP=1
TOTAL_STEPS=1

# Calculate total steps
if [ "$TARGET" = "all" ] || [ "$TARGET" = "ultra-all" ]; then
    TOTAL_STEPS=4
elif [ "$TARGET" = "test" ] || [ "$TARGET" = "ultra-test" ]; then
    TOTAL_STEPS=2
elif [ "$TARGET" = "dev" ] || [ "$TARGET" = "ultra-dev" ]; then
    TOTAL_STEPS=1
elif [ "$TARGET" = "prod" ] || [ "$TARGET" = "ultra-prod" ]; then
    TOTAL_STEPS=1
fi

# 1. Build Test Stage (always needed for test target or all)
if [ "$TARGET" = "test" ] || [ "$TARGET" = "all" ] || [ "$TARGET" = "ultra-test" ] || [ "$TARGET" = "ultra-all" ]; then
    echo -e "${B_PURPLE}[$STEP/$TOTAL_STEPS]${RESET} ${P_CYAN}Building test image...${RESET}"
    if docker build --target test -t "${IMAGE_NAME}:test" .; then
        echo -e "   ${B_GREEN}+ Test image built:${RESET} ${B_WHITE}${IMAGE_NAME}:test${RESET}"
    else
        echo -e "   ${B_RED}x Test build failed${RESET}"
        exit 1
    fi
    STEP=$((STEP + 1))
fi

# 2. Run Tests (only for test target or all)
if [ "$TARGET" = "test" ] || [ "$TARGET" = "all" ] || [ "$TARGET" = "ultra-test" ] || [ "$TARGET" = "ultra-all" ]; then
    echo -e "${B_PURPLE}[$STEP/$TOTAL_STEPS]${RESET} ${P_CYAN}Running tests...${RESET}"
    if docker run --rm "${IMAGE_NAME}:test"; then
        echo -e "   ${B_GREEN}+ Tests passed${RESET}"
    else
        echo -e "   ${B_RED}x Tests failed${RESET}"
        exit 1
    fi
    STEP=$((STEP + 1))
fi

# 3. Build Development Stage
if [ "$TARGET" = "dev" ] || [ "$TARGET" = "all" ]; then
    echo -e "${B_PURPLE}[$STEP/$TOTAL_STEPS]${RESET} ${P_CYAN}Building development image...${RESET}"
    if docker build --target dev -t "${IMAGE_NAME}:dev" .; then
        echo -e "   ${B_GREEN}+ Development image built:${RESET} ${B_WHITE}${IMAGE_NAME}:dev${RESET}"
    else
        echo -e "   ${B_RED}x Development build failed${RESET}"
        exit 1
    fi
    STEP=$((STEP + 1))
fi

# 3. Build Ultra Development Stage
if [ "$TARGET" = "ultra-dev" ] || [ "$TARGET" = "ultra-all" ]; then
    echo -e "${B_PURPLE}[$STEP/$TOTAL_STEPS]${RESET} ${P_CYAN}Building ultra development image...${RESET}"
    if docker build --target ultra-dev -t "${IMAGE_NAME}:dev" .; then
        echo -e "   ${B_GREEN}+ Ultra development image built:${RESET} ${B_WHITE}${IMAGE_NAME}:dev${RESET}"
    else
        echo -e "   ${B_RED}x Ultra development build failed${RESET}"
        exit 1
    fi
    STEP=$((STEP + 1))
fi

# 4. Build Production Stage
if [ "$TARGET" = "prod" ] || [ "$TARGET" = "all" ]; then
    echo -e "${B_PURPLE}[$STEP/$TOTAL_STEPS]${RESET} ${P_CYAN}Building production image...${RESET}"
    if docker build --target prod -t "${IMAGE_NAME}:latest" .; then
        echo -e "   ${B_GREEN}+ Production image built:${RESET} ${B_WHITE}${IMAGE_NAME}:latest${RESET}"
    else
        echo -e "   ${B_RED}x Production build failed${RESET}"
        exit 1
    fi
    STEP=$((STEP + 1))
fi

# 4. Build Slim Production Stage
if [ "$TARGET" = "prod-slim" ]; then
    echo -e "${B_PURPLE}[$STEP/$TOTAL_STEPS]${RESET} ${P_CYAN}Building slim production image...${RESET}"
    if docker build --target prod-slim -t "${IMAGE_NAME}:slim" .; then
        echo -e "   ${B_GREEN}+ Slim production image built:${RESET} ${B_WHITE}${IMAGE_NAME}:slim${RESET}"
    else
        echo -e "   ${B_RED}x Slim production build failed${RESET}"
        exit 1
    fi
    STEP=$((STEP + 1))
fi

# 4. Build Minimal Production Stage
if [ "$TARGET" = "prod-minimal" ]; then
    echo -e "${B_PURPLE}[$STEP/$TOTAL_STEPS]${RESET} ${P_CYAN}Building minimal production image...${RESET}"
    if docker build --target prod-minimal -t "${IMAGE_NAME}:minimal" .; then
        echo -e "   ${B_GREEN}+ Minimal production image built:${RESET} ${B_WHITE}${IMAGE_NAME}:minimal${RESET}"
    else
        echo -e "   ${B_RED}x Minimal production build failed${RESET}"
        exit 1
    fi
    STEP=$((STEP + 1))
fi

# 4. Build Ultra Production Stage (for ultra-all)
if [ "$TARGET" = "ultra-all" ]; then
    echo -e "${B_PURPLE}[$STEP/$TOTAL_STEPS]${RESET} ${P_CYAN}Building ultra production image...${RESET}"
    if docker build --target ultra-prod -t "${IMAGE_NAME}:ultra" .; then
        echo -e "   ${B_GREEN}+ Ultra production image built:${RESET} ${B_WHITE}${IMAGE_NAME}:ultra${RESET}"
    else
        echo -e "   ${B_RED}x Ultra production build failed${RESET}"
        exit 1
    fi
    STEP=$((STEP + 1))
fi

# 5. Build Ultra-Production Stage (ultra slim)
if [ "$TARGET" = "ultra-prod" ]; then
    echo -e "${B_PURPLE}[$STEP/$TOTAL_STEPS]${RESET} ${P_CYAN}Building ultra-production image...${RESET}"
    if docker build --target ultra-prod -t "${IMAGE_NAME}:ultra" .; then
        echo -e "   ${B_GREEN}+ Ultra-production image built:${RESET} ${B_WHITE}${IMAGE_NAME}:ultra${RESET}"
    else
        echo -e "   ${B_RED}x Ultra-production build failed${RESET}"
        exit 1
    fi
    STEP=$((STEP + 1))
fi

# Optional: Clean up intermediate images
echo -e "${B_PURPLE}[Bonus]${RESET} ${P_CYAN}Cleaning up dangling images...${RESET}"
if docker image prune -f &>/dev/null; then
    echo -e "   ${B_GREEN}+ Cleanup completed${RESET}"
else
    echo -e "   ${GRAY}- Cleanup skipped${RESET}"
fi

echo ""
echo -e "${GRAY}======================================================================${RESET}"
echo -e " ${B_PURPLE}SUCCESS!${RESET} ${P_GREEN}Docker images ready:${RESET}"

# Show only the images that were actually built
if [ "$TARGET" = "test" ] || [ "$TARGET" = "all" ] || [ "$TARGET" = "ultra-test" ] || [ "$TARGET" = "ultra-all" ]; then
    echo -e "   ${B_WHITE}${IMAGE_NAME}:test${RESET}     ${GRAY}(testing)${RESET}"
fi
if [ "$TARGET" = "dev" ] || [ "$TARGET" = "all" ] || [ "$TARGET" = "ultra-dev" ] || [ "$TARGET" = "ultra-all" ]; then
    echo -e "   ${B_WHITE}${IMAGE_NAME}:dev${RESET}      ${GRAY}(development)${RESET}"
fi
if [ "$TARGET" = "prod" ] || [ "$TARGET" = "all" ]; then
    echo -e "   ${B_WHITE}${IMAGE_NAME}:latest${RESET}  ${GRAY}(production)${RESET}"
fi
if [ "$TARGET" = "prod-slim" ]; then
    echo -e "   ${B_WHITE}${IMAGE_NAME}:slim${RESET}    ${GRAY}(slim-production)${RESET}"
fi
if [ "$TARGET" = "prod-minimal" ]; then
    echo -e "   ${B_WHITE}${IMAGE_NAME}:minimal${RESET} ${GRAY}(minimal-production)${RESET}"
fi
if [ "$TARGET" = "ultra-prod" ]; then
    echo -e "   ${B_WHITE}${IMAGE_NAME}:ultra${RESET}    ${GRAY}(ultra-production)${RESET}"
fi
if [ "$TARGET" = "ultra-all" ]; then
    echo -e "   ${B_WHITE}${IMAGE_NAME}:ultra${RESET}    ${GRAY}(ultra-production)${RESET}"
fi

echo -e "${GRAY}======================================================================${RESET}"
echo ""
echo -e "${B_CYAN}Usage examples:${RESET}"

# Show relevant usage examples based on what was built
if [ "$TARGET" = "dev" ] || [ "$TARGET" = "all" ]; then
    echo -e "  ${GRAY}# Run development container${RESET}"
    echo -e "  ${B_WHITE}docker run -p 8080:8080 -p 9090:9090 ${IMAGE_NAME}:dev${RESET}"
    echo ""
fi

if [ "$TARGET" = "prod" ] || [ "$TARGET" = "all" ]; then
    echo -e "  ${GRAY}# Run production container${RESET}"
    echo -e "  ${B_WHITE}docker run -p 8080:8080 -p 9090:9090 ${IMAGE_NAME}:latest${RESET}"
    echo ""
fi

if [ "$TARGET" = "test" ] || [ "$TARGET" = "all" ]; then
    echo -e "  ${GRAY}# Run tests${RESET}"
    echo -e "  ${B_WHITE}docker run --rm ${IMAGE_NAME}:test${RESET}"
fi
