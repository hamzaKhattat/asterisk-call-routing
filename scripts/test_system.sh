#!/bin/bash

echo "Testing Asterisk Call Routing System..."

# Test health endpoint
echo -n "Testing health endpoint: "
HEALTH=$(curl -s http://localhost:8000/health)
if [ "$HEALTH" == "OK" ]; then
   echo "✓ OK"
else
   echo "✗ Failed"
   exit 1
fi

# Test stats endpoint
echo -n "Testing stats endpoint: "
STATS=$(curl -s http://localhost:8000/stats)
if [ $? -eq 0 ]; then
   echo "✓ OK"
   echo "Stats: $STATS" | jq '.'
else
   echo "✗ Failed"
fi

# Test API health check
echo -n "Testing API health check: "
API_HEALTH=$(curl -s http://localhost:8000/api/health)
if [ $? -eq 0 ]; then
   echo "✓ OK"
else
   echo "✗ Failed"
fi

# Test incoming call processing
echo -n "Testing incoming call processing: "
RESPONSE=$(curl -s -X POST http://localhost:8000/process-incoming \
   -d "uniqueid=test_$(date +%s)" \
   -d "callerid=19543004835" \
   -d "extension=50764137984")

if [[ "$RESPONSE" == *"SET VARIABLE"* ]]; then
   echo "✓ OK"
   echo "Response: $RESPONSE"
else
   echo "✗ Failed"
   echo "Response: $RESPONSE"
fi

echo "Test complete!"
