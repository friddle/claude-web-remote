#!/bin/bash

# ClauDED Android Build Script
# This script builds a signed Android APK for the ClauDED mobile client

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="clauded"
PACKAGE_NAME="com.friddle.clauded"
KEYSTORE_NAME="clauded-release.keystore"
KEY_ALIAS="clauded-key"
OUTPUT_DIR="$HOME/.ssh/android"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ANDROID_DIR="$PROJECT_ROOT/mobile_client/android"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}ClauDED Android Build Script${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check if we're in the right directory
if [ ! -d "$ANDROID_DIR" ]; then
    echo -e "${RED}Error: Android directory not found at $ANDROID_DIR${NC}"
    exit 1
fi

# Step 1: Generate keystore if it doesn't exist
echo -e "${YELLOW}[1/5] Checking signing keystore...${NC}"
KEYSTORE_PATH="$ANDROID_DIR/app/$KEYSTORE_NAME"

if [ ! -f "$KEYSTORE_PATH" ]; then
    echo -e "${YELLOW}Generating new signing keystore...${NC}"

    # Generate keystore with default password
    keytool -genkeypair \
        -v \
        -storepass "clauded123" \
        -keypass "clauded123" \
        -keystore "$KEYSTORE_PATH" \
        -alias "$KEY_ALIAS" \
        -keyalg RSA \
        -keysize 2048 \
        -validity 10000 \
        -dname "CN=ClauDED, OU=Development, O=Friddle, L=City, ST=State, C=US"

    echo -e "${GREEN}✓ Keystore generated at $KEYSTORE_PATH${NC}"
else
    echo -e "${GREEN}✓ Keystore already exists at $KEYSTORE_PATH${NC}"
fi

echo ""

# Step 2: Create signing config if it doesn't exist
echo -e "${YELLOW}[2/5] Configuring signing...${NC}"
SIGNING_CONFIG="$ANDROID_DIR/app/signing-config.properties"

if [ ! -f "$SIGNING_CONFIG" ]; then
    cat > "$SIGNING_CONFIG" << EOF
STORE_FILE=$KEYSTORE_NAME
STORE_PASSWORD=clauded123
KEY_ALIAS=$KEY_ALIAS
KEY_PASSWORD=clauded123
EOF
    echo -e "${GREEN}✓ Signing config created${NC}"
else
    echo -e "${GREEN}✓ Signing config already exists${NC}"
fi

echo ""

# Step 3: Sync Capacitor
echo -e "${YELLOW}[3/5] Syncing Capacitor project...${NC}"
cd "$PROJECT_ROOT/mobile_client"
npx cap sync android

# Fix Java version to 17 after sync (Capacitor defaults to 21)
echo -e "${YELLOW}Fixing Java version compatibility...${NC}"
sed -i '' 's/JavaVersion\.VERSION_21/JavaVersion.VERSION_17/g' "$ANDROID_DIR/app/capacitor.build.gradle"
sed -i '' 's/JavaVersion\.VERSION_21/JavaVersion.VERSION_17/g' "$PROJECT_ROOT/mobile_client/node_modules/@capacitor/android/capacitor/build.gradle"
sed -i '' 's/JavaVersion\.VERSION_21/JavaVersion.VERSION_17/g' "$PROJECT_ROOT/mobile_client/node_modules/@capacitor/app/android/build.gradle"
sed -i '' 's/JavaVersion\.VERSION_21/JavaVersion.VERSION_17/g' "$PROJECT_ROOT/mobile_client/node_modules/@capacitor/local-notifications/android/build.gradle"
echo -e "${GREEN}✓ Capacitor sync completed${NC}"

echo ""

# Step 4: Build APK
echo -e "${YELLOW}[4/5] Building release APK...${NC}"
cd "$ANDROID_DIR"
./gradlew assembleRelease
echo -e "${GREEN}✓ APK build completed${NC}"

echo ""

# Step 5: Copy artifacts
echo -e "${YELLOW}[5/5] Copying artifacts to $OUTPUT_DIR...${NC}"
mkdir -p "$OUTPUT_DIR"

# Find and copy the latest APK
APK_PATH=$(find "$ANDROID_DIR/app/build/outputs/apk/release" -name "*.apk" | head -1)
if [ -z "$APK_PATH" ]; then
    echo -e "${RED}Error: APK not found in build output${NC}"
    exit 1
fi

APK_NAME="${APP_NAME}-$(date +%Y%m%d-%H%M%S).apk"
cp "$APK_PATH" "$OUTPUT_DIR/$APK_NAME"
echo -e "${GREEN}✓ APK copied to $OUTPUT_DIR/$APK_NAME${NC}"

# Copy keystore to output directory
cp "$KEYSTORE_PATH" "$OUTPUT_DIR/$KEYSTORE_NAME"
echo -e "${GREEN}✓ Keystore copied to $OUTPUT_DIR/$KEYSTORE_NAME${NC}"

# Create info file
cat > "$OUTPUT_DIR/${APP_NAME}-build-info.txt" << EOF
ClauDED Android Build Information
===================================
Build Date: $(date)
APK File: $APK_NAME
Keystore File: $KEYSTORE_NAME
Package: $PACKAGE_NAME
Key Alias: $KEY_ALIAS
Store Password: clauded123
Key Password: clauded123
EOF

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "APK: ${GREEN}$OUTPUT_DIR/$APK_NAME${NC}"
echo -e "Keystore: ${GREEN}$OUTPUT_DIR/$KEYSTORE_NAME${NC}"
echo ""
echo -e "${YELLOW}Warning: Keep your keystore and passwords secure!${NC}"
echo ""
