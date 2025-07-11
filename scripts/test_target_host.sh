#!/bin/bash
# This script tests the set_target_host.sh script

# Set up test environment
TEST_DIR=$(mktemp -d)
TEST_ENV_FILE="$TEST_DIR/.env"
echo "🧪 Testing in temporary directory: $TEST_DIR"

# Copy the script to the test directory
cp ./set_target_host.sh "$TEST_DIR/"
cd "$TEST_DIR"

# Test 1: Fresh install (no .env file)
echo "🧪 Test 1: Fresh install (no .env file)"
IP=$(./set_target_host.sh ./.env)

# Check if the .env file was created
if [ -f .env ]; then
    echo "✅ .env file created successfully"
    echo "📄 Contents of .env:"
    cat .env
else
    echo "❌ .env file not created"
fi

# Check if TARGET_HOST was set
if [ -n "$IP" ]; then
    echo "✅ TARGET_HOST set to: $IP"
else
    echo "❌ TARGET_HOST not set"
fi

# Test if the IP is valid
if [[ $IP =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]] || [ "$IP" = "localhost" ]; then
    echo "✅ IP format is valid"
else
    echo "❌ Invalid IP format: $IP"
fi

# Test 2: Existing .env file with different TARGET_HOST
echo -e "\n🧪 Test 2: Existing .env file with different TARGET_HOST"
echo "TARGET_HOST=192.168.1.100" > .env
echo "SOME_OTHER_VAR=test_value" >> .env
echo "📄 Original .env contents:"
cat .env

# Run the script again
IP=$(./set_target_host.sh ./.env)

# Check if the .env file was updated correctly
echo "📄 Updated .env contents:"
cat .env

# Check if TARGET_HOST was updated but SOME_OTHER_VAR was preserved
if grep -q "TARGET_HOST=$IP" .env && grep -q "SOME_OTHER_VAR=test_value" .env; then
    echo "✅ TARGET_HOST updated and other variables preserved"
else
    echo "❌ .env file not updated correctly"
fi

# Test 3: Existing .env file without TARGET_HOST
echo -e "\n🧪 Test 3: Existing .env file without TARGET_HOST"
echo "SOME_OTHER_VAR=test_value" > .env
echo "📄 Original .env contents:"
cat .env

# Run the script again
IP=$(./set_target_host.sh ./.env)

# Check if the .env file was updated correctly
echo "📄 Updated .env contents:"
cat .env

# Check if TARGET_HOST was added and SOME_OTHER_VAR was preserved
if grep -q "TARGET_HOST=$IP" .env && grep -q "SOME_OTHER_VAR=test_value" .env; then
    echo "✅ TARGET_HOST added and other variables preserved"
else
    echo "❌ .env file not updated correctly"
fi

# Clean up
cd - > /dev/null
rm -rf "$TEST_DIR"
echo -e "\n🧹 Cleaned up test directory"
echo "✅ All tests completed"