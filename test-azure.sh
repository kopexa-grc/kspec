#!/bin/bash

# Azure Provider Test Script
# This script tests the Azure provider with minimal configuration

set -e

echo "🔍 Azure Provider Test Script"
echo "=============================="

# Check if credentials are set
if [ -z "$AZURE_SUBSCRIPTION_ID" ]; then
    echo "❌ Error: AZURE_SUBSCRIPTION_ID not set"
    echo ""
    echo "Please set Azure credentials:"
    echo "  export AZURE_SUBSCRIPTION_ID=<your-subscription-id>"
    echo "  export AZURE_TENANT_ID=<your-tenant-id>"
    echo "  export AZURE_CLIENT_ID=<your-client-id>"
    echo "  export AZURE_CLIENT_SECRET=<your-client-secret>"
    echo ""
    echo "Or use Azure CLI:"
    echo "  az login"
    exit 1
fi

# Build kspec
echo ""
echo "📦 Building kspec..."
go build -o kspec ./cmd/kspec

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Test 1: Quick connectivity test with --help
echo ""
echo "🧪 Test 1: Checking kspec CLI..."
./kspec --version 2>/dev/null || ./kspec --help | head -5

# Test 2: Run minimal Azure scan
echo ""
echo "🧪 Test 2: Running Azure security scan..."
echo "   Subscription: $AZURE_SUBSCRIPTION_ID"

./kspec scan azure subscription $AZURE_SUBSCRIPTION_ID \
  --credential-type env \
  -f policies/azure-security.yml

SCAN_RESULT=$?

echo ""
if [ $SCAN_RESULT -eq 0 ]; then
    echo "✅ Azure scan completed successfully!"
else
    echo "⚠️  Scan completed with errors (exit code: $SCAN_RESULT)"
fi

echo ""
echo "=============================="
echo "Test complete!"
