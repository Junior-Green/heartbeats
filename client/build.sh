
PROJECT_DIR="/Users/juniorgreen/Documents/heartbeats/client"
GO_PROJECT_DIR="/Users/juniorgreen/Documents/heartbeats/backend"
OUTPUT_DIR="${PROJECT_DIR}/GeneratedBinaries"
CGO_ENABLED=1
CGO_CFLAGS="-mmacosx-version-min=$MACOSX_DEPLOYMENT_TARGET"
CGO_LDFLAGS="-mmacosx-version-min=$MACOSX_DEPLOYMENT_TARGET"
BUNDLE_ID="com.heartbeats"

# Ensure the output directory exists
mkdir -p "$OUTPUT_DIR"

# Compile for Intel (amd64)
GO111MODULE=auto GOOS=darwin GOARCH=amd64 go build -C $GO_PROJECT_DIR -ldflags="-w -s" -o "$OUTPUT_DIR/$BUNDLE_ID.darwin_amd64" "."

# Compile for Apple Silicon (arm64)
GO111MODULE=auto GOOS=darwin GOARCH=arm64 go build -C $GO_PROJECT_DIR -ldflags="-w -s" -o "$OUTPUT_DIR/$BUNDLE_ID.darwin_arm64" "."

# Create universal binary
lipo -create -output "$OUTPUT_DIR/$BUNDLE_ID.universal" \
    "$OUTPUT_DIR/$BUNDLE_ID.darwin_amd64" "$OUTPUT_DIR/$BUNDLE_ID.darwin_arm64"

chmod +x "$OUTPUT_DIR/darwin_amd64"
chmod +x "$OUTPUT_DIR/darwin_arm64"
chmod +x "$OUTPUT_DIR/universal"

echo "Universal binary created at $OUTPUT_DIR/$BUNDLE_ID.universal"
