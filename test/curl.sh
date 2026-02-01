#!/bin/bash

# Available endpoints and their corresponding request files
ENDPOINTS=("bigwig" "bigbed" "transcript" "browser")
REQUEST_FILES=("bigWigRequest.json" "bigBedRequest.json" "transcriptRequest.json" "browserRequest.json")

# Base URLs
DEV_URL="http://localhost:8080"
PROD_URL="https://genome-browser-api.fly.dev"

# Script directory (to find request files)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Usage function
usage() {
    echo "Usage: $0 <endpoint> [dev|prod]"
    echo ""
    echo "Endpoints:"
    for ep in "${ENDPOINTS[@]}"; do
        echo "  - $ep"
    done
    echo ""
    echo "Environment:"
    echo "  - dev  (default): $DEV_URL"
    echo "  - prod:           $PROD_URL"
    echo ""
    echo "Examples:"
    echo "  $0 bigwig"
    echo "  $0 bigwig dev"
    echo "  $0 bigwig prod"
    exit 1
}

# Check if endpoint is provided
if [ -z "$1" ]; then
    usage
fi

ENDPOINT="$1"
ENV="${2:-dev}"

# Set base URL based on environment
if [ "$ENV" = "prod" ]; then
    BASE_URL="$PROD_URL"
elif [ "$ENV" = "dev" ]; then
    BASE_URL="$DEV_URL"
else
    echo "Error: Invalid environment '$ENV'. Use 'dev' or 'prod'."
    exit 1
fi

# Find the matching endpoint
REQUEST_FILE=""
for i in "${!ENDPOINTS[@]}"; do
    if [ "${ENDPOINTS[$i]}" = "$ENDPOINT" ]; then
        REQUEST_FILE="${REQUEST_FILES[$i]}"
        break
    fi
done

# Check if endpoint was found
if [ -z "$REQUEST_FILE" ]; then
    echo "Error: Invalid endpoint '$ENDPOINT'."
    echo ""
    echo "Available endpoints:"
    for ep in "${ENDPOINTS[@]}"; do
        echo "  - $ep"
    done
    exit 1
fi

# Execute curl request
curl -s -X POST -H "Content-Type: application/json" \
    -d @"$SCRIPT_DIR/request/$REQUEST_FILE" \
    "$BASE_URL/v1/$ENDPOINT" | jq .
