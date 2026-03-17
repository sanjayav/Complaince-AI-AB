#!/bin/bash

# Script to generate an API key for frontend developers
# This script calls the admin endpoint to create a new API key

set -e

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
ADMIN_API_KEY="${ADMIN_API_KEY:-admin_key_placeholder}"  # You'll need to set this

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔑 JLR Document Intelligence - Frontend API Key Generator${NC}"
echo "=================================================="

# Check if required tools are available
if ! command -v curl &> /dev/null; then
    echo -e "${RED}❌ Error: curl is required but not installed${NC}"
    exit 1
fi

# Check if jq is available for JSON parsing
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}⚠️  Warning: jq is not installed. JSON output will not be formatted.${NC}"
    JQ_AVAILABLE=false
else
    JQ_AVAILABLE=true
fi

# Function to generate API key
generate_api_key() {
    local developer_name="$1"
    local permissions="${2:-default}"
    
    echo -e "${BLUE}📝 Generating API key for: ${developer_name}${NC}"
    echo -e "${BLUE}🔐 Permissions: ${permissions}${NC}"
    
    # Set permissions based on input
    local permission_array
    if [ "$permissions" = "default" ]; then
        permission_array='["documents:read","documents:upload","search:semantic","rag:ask","answers:read","health:read"]'
    elif [ "$permissions" = "admin" ]; then
        permission_array='["documents:read","documents:upload","documents:delete","search:semantic","rag:ask","answers:read","answers:write","health:read","metrics:read","admin:manage"]'
    else
        permission_array='["documents:read","documents:upload","search:semantic","rag:ask","answers:read","health:read"]'
    fi
    
    # Generate the API key
    local response
    if [ "$JQ_AVAILABLE" = true ]; then
        response=$(curl -s -X POST "${API_BASE_URL}/v1/admin/apikeys" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer ${ADMIN_API_KEY}" \
            -d "{\"name\":\"${developer_name}\",\"permissions\":${permission_array}}" | jq '.')
    else
        response=$(curl -s -X POST "${API_BASE_URL}/v1/admin/apikeys" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer ${ADMIN_API_KEY}" \
            -d "{\"name\":\"${developer_name}\",\"permissions\":${permission_array}}")
    fi
    
    # Check if the request was successful
    if echo "$response" | grep -q "error"; then
        echo -e "${RED}❌ Error generating API key:${NC}"
        echo "$response"
        return 1
    fi
    
    # Extract the API key from response
    local api_key
    if [ "$JQ_AVAILABLE" = true ]; then
        api_key=$(echo "$response" | jq -r '.api_key.key')
    else
        api_key=$(echo "$response" | echo "$response" | sed -n 's/.*"key":"\([^"]*\)".*/\1/p')
    fi
    
    if [ "$api_key" = "null" ] || [ -z "$api_key" ]; then
        echo -e "${RED}❌ Failed to extract API key from response${NC}"
        echo "$response"
        return 1
    fi
    
    echo -e "${GREEN}✅ API key generated successfully!${NC}"
    echo ""
    echo -e "${YELLOW}🔑 API Key: ${api_key}${NC}"
    echo ""
    echo -e "${BLUE}📋 Usage Instructions:${NC}"
    echo "1. Include this API key in your requests using one of these headers:"
    echo "   - X-API-Key: ${api_key}"
    echo "   - Authorization: Bearer ${api_key}"
    echo "   - Authorization: ApiKey ${api_key}"
    echo ""
    echo -e "${YELLOW}⚠️  IMPORTANT: Save this key securely! It won't be shown again.${NC}"
    echo ""
    
    # Save to file if requested
    read -p "Save API key to file? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        local filename="api_key_${developer_name}_$(date +%Y%m%d_%H%M%S).txt"
        cat > "$filename" << EOF
JLR Document Intelligence - API Key
===================================
Developer: ${developer_name}
Generated: $(date)
API Key: ${api_key}

Usage:
- Header: X-API-Key: ${api_key}
- Header: Authorization: Bearer ${api_key}
- Header: Authorization: ApiKey ${api_key}

⚠️  Keep this key secure and don't share it publicly!
EOF
        echo -e "${GREEN}✅ API key saved to: ${filename}${NC}"
    fi
    
    return 0
}

# Function to list existing API keys
list_api_keys() {
    echo -e "${BLUE}📋 Listing existing API keys...${NC}"
    
    local response
    if [ "$JQ_AVAILABLE" = true ]; then
        response=$(curl -s -X GET "${API_BASE_URL}/v1/admin/apikeys" \
            -H "Authorization: Bearer ${ADMIN_API_KEY}" | jq '.')
    else
        response=$(curl -s -X GET "${API_BASE_URL}/v1/admin/apikeys" \
            -H "Authorization: Bearer ${ADMIN_API_KEY}")
    fi
    
    if echo "$response" | grep -q "error"; then
        echo -e "${RED}❌ Error listing API keys:${NC}"
        echo "$response"
        return 1
    fi
    
    echo "$response"
}

# Function to revoke an API key
revoke_api_key() {
    local key_hash="$1"
    
    if [ -z "$key_hash" ]; then
        echo -e "${RED}❌ Error: key_hash is required${NC}"
        echo "Usage: $0 revoke <key_hash>"
        return 1
    fi
    
    echo -e "${BLUE}🗑️  Revoking API key: ${key_hash}${NC}"
    
    local response
    if [ "$JQ_AVAILABLE" = true ]; then
        response=$(curl -s -X DELETE "${API_BASE_URL}/v1/admin/apikeys?key_hash=${key_hash}" \
            -H "Authorization: Bearer ${ADMIN_API_KEY}" | jq '.')
    else
        response=$(curl -s -X DELETE "${API_BASE_URL}/v1/admin/apikeys?key_hash=${key_hash}" \
            -H "Authorization: Bearer ${ADMIN_API_KEY}")
    fi
    
    if echo "$response" | grep -q "error"; then
        echo -e "${RED}❌ Error revoking API key:${NC}"
        echo "$response"
        return 1
    fi
    
    echo -e "${GREEN}✅ API key revoked successfully${NC}"
    echo "$response"
}

# Function to show usage guide
show_usage_guide() {
    echo -e "${BLUE}📚 API Usage Guide${NC}"
    echo "=================="
    
    local response
    if [ "$JQ_AVAILABLE" = true ]; then
        response=$(curl -s -X GET "${API_BASE_URL}/v1/apikey/guide" \
            -H "X-API-Key: ${ADMIN_API_KEY}" | jq '.')
    else
        response=$(curl -s -X GET "${API_BASE_URL}/v1/apikey/guide" \
            -H "X-API-Key: ${ADMIN_API_KEY}")
    fi
    
    if echo "$response" | grep -q "error"; then
        echo -e "${RED}❌ Error fetching usage guide:${NC}"
        echo "$response"
        return 1
    fi
    
    echo "$response"
}

# Main script logic
case "${1:-help}" in
    "generate"|"gen")
        if [ -z "$2" ]; then
            echo -e "${RED}❌ Error: Developer name is required${NC}"
            echo "Usage: $0 generate <developer_name> [permissions]"
            echo "Permissions: default, admin"
            exit 1
        fi
        generate_api_key "$2" "${3:-default}"
        ;;
    "list"|"ls")
        list_api_keys
        ;;
    "revoke"|"rm")
        revoke_api_key "$2"
        ;;
    "guide"|"help"|"usage")
        show_usage_guide
        ;;
    *)
        echo -e "${BLUE}🔑 JLR Document Intelligence - Frontend API Key Generator${NC}"
        echo ""
        echo "Usage:"
        echo "  $0 generate <developer_name> [permissions]  - Generate new API key"
        echo "  $0 list                                    - List existing API keys"
        echo "  $0 revoke <key_hash>                       - Revoke an API key"
        echo "  $0 guide                                   - Show usage guide"
        echo "  $0 help                                    - Show this help"
        echo ""
        echo "Examples:"
        echo "  $0 generate john_doe                       - Generate key with default permissions"
        echo "  $0 generate jane_admin admin               - Generate key with admin permissions"
        echo "  $0 list                                    - List all API keys"
        echo "  $0 revoke abc123def456                     - Revoke specific key"
        echo ""
        echo "Environment Variables:"
        echo "  API_BASE_URL     - API base URL (default: http://localhost:8080)"
        echo "  ADMIN_API_KEY    - Admin API key for authentication"
        echo ""
        echo "Note: Make sure the JLRDI backend is running and accessible."
        ;;
esac
