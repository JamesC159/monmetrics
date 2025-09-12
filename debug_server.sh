#!/bin/bash

echo "🔍 MonMetrics Server Diagnostic"
echo "==============================="

# Test health endpoint first
echo "1. Testing Health Endpoint..."
HEALTH_RESPONSE=$(curl -s -w "HTTP_STATUS:%{http_code}" "http://localhost:8080/health")
HTTP_STATUS=$(echo "$HEALTH_RESPONSE" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
RESPONSE_BODY=$(echo "$HEALTH_RESPONSE" | sed 's/HTTP_STATUS:[0-9]*$//')

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ Health check passed: $RESPONSE_BODY"
else
    echo "❌ Health check failed with status: $HTTP_STATUS"
    echo "Response: $RESPONSE_BODY"
    echo "❌ Server may not be running. Please run 'make dev' first."
    exit 1
fi

# Test search endpoint (the problematic one)
echo ""
echo "2. Testing Search Endpoint..."
SEARCH_RESPONSE=$(curl -s -w "HTTP_STATUS:%{http_code}" "http://localhost:8080/api/cards/search?q=test")
SEARCH_STATUS=$(echo "$SEARCH_RESPONSE" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
SEARCH_BODY=$(echo "$SEARCH_RESPONSE" | sed 's/HTTP_STATUS:[0-9]*$//')

if [ "$SEARCH_STATUS" = "200" ]; then
    echo "✅ Search endpoint working: $SEARCH_BODY"
elif [ "$SEARCH_STATUS" = "500" ]; then
    echo "❌ Search endpoint returning 500 error:"
    echo "$SEARCH_BODY"
    echo "This confirms the handler issue we need to fix."
elif [ "$SEARCH_STATUS" = "404" ]; then
    echo "❌ Search endpoint not found (404)"
    echo "Routing issue - endpoint not properly registered"
else
    echo "❌ Search endpoint failed with status: $SEARCH_STATUS"
    echo "Response: $SEARCH_BODY"
fi

# Test CORS with OPTIONS request
echo ""
echo "3. Testing CORS Configuration..."
CORS_RESPONSE=$(curl -s -w "HTTP_STATUS:%{http_code}" -X OPTIONS \
    -H "Origin: http://localhost:3000" \
    -H "Access-Control-Request-Method: GET" \
    -H "Access-Control-Request-Headers: Content-Type" \
    "http://localhost:8080/api/cards/search")

CORS_STATUS=$(echo "$CORS_RESPONSE" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)

if [ "$CORS_STATUS" = "200" ]; then
    echo "✅ CORS preflight request successful"
else
    echo "❌ CORS preflight failed with status: $CORS_STATUS"
fi

# Check if MongoDB is accessible
echo ""
echo "4. Testing Database Connection..."
if command -v docker &> /dev/null; then
    if docker ps | grep -q "mongo"; then
        echo "✅ MongoDB container is running"

        # Test MongoDB connection
        DB_TEST=$(docker exec -it $(docker ps -q -f name=mongo) mongosh --eval "db.adminCommand('ping')" --quiet 2>/dev/null)
        if echo "$DB_TEST" | grep -q "ok"; then
            echo "✅ MongoDB is responding"
        else
            echo "❌ MongoDB is not responding properly"
        fi
    else
        echo "❌ MongoDB container not found. Run 'make setup' to start MongoDB."
    fi
else
    echo "⚠️  Docker not found, cannot check MongoDB status"
fi

# Check if data exists
echo ""
echo "5. Testing Database Data..."
if command -v docker &> /dev/null && docker ps | grep -q "mongo"; then
    CARD_COUNT=$(docker exec $(docker ps -q -f name=mongo) mongosh monmetrics --eval "db.cards.countDocuments({})" --quiet 2>/dev/null | tail -n1)
    if [ "$CARD_COUNT" -gt 0 ] 2>/dev/null; then
        echo "✅ Database has $CARD_COUNT cards"
    else
        echo "❌ Database is empty. Run 'make seed' to populate data."
    fi
fi

echo ""
echo "📋 Summary:"
echo "==========="
echo "Health Status: $([ "$HTTP_STATUS" = "200" ] && echo "✅ OK" || echo "❌ FAILED")"
echo "Search Status: $([ "$SEARCH_STATUS" = "200" ] && echo "✅ OK" || echo "❌ FAILED")"
echo "CORS Status: $([ "$CORS_STATUS" = "200" ] && echo "✅ OK" || echo "❌ FAILED")"

if [ "$SEARCH_STATUS" != "200" ]; then
    echo ""
    echo "🔧 Recommended Actions:"
    echo "1. Replace the current handlers.go with the fixed version"
    echo "2. Restart the server: Ctrl+C then 'make dev'"
    echo "3. If still failing, try: 'make reset && make full-setup'"
fi

echo ""
echo "🌐 Test URLs:"
echo "Health: http://localhost:8080/health"
echo "Search: http://localhost:8080/api/cards/search?q=charizard"
echo "Frontend: http://localhost:3000"