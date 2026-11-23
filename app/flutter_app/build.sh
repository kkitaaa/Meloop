#!/bin/bash
# Flutter build script for different environments
# Usage: ./build.sh [dev|staging|prod] [android|windows|ios]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Default values
ENVIRONMENT="${1:-dev}"
PLATFORM="${2:-android}"

# Configuration mapping
case "$ENVIRONMENT" in
  dev|development)
    FLUTTER_ENV="development"
    API_BASE_URL="${API_BASE_URL:-}"
    K8S_API_GATEWAY_URL=""
    K8S_INGRESS_HOST=""
    BUILD_NAME="1.0.0"
    BUILD_NUMBER="1"
    ;;
  staging)
    FLUTTER_ENV="staging"
    API_BASE_URL="${API_BASE_URL:-}"
    K8S_API_GATEWAY_URL="${K8S_API_GATEWAY_URL_STAGING:-}"
    K8S_INGRESS_HOST="${K8S_INGRESS_HOST_STAGING:-}"
    if [ -z "$API_BASE_URL" ] && [ -z "$K8S_API_GATEWAY_URL" ] && [ -z "$K8S_INGRESS_HOST" ]; then
      echo "Set API_BASE_URL, K8S_API_GATEWAY_URL_STAGING, or K8S_INGRESS_HOST_STAGING."
      exit 1
    fi
    BUILD_NAME="1.0.0"
    BUILD_NUMBER="2"
    ;;
  prod|production)
    FLUTTER_ENV="production"
    API_BASE_URL="${API_BASE_URL:-}"
    K8S_API_GATEWAY_URL="${K8S_API_GATEWAY_URL_PROD:-}"
    K8S_INGRESS_HOST="${K8S_INGRESS_HOST_PROD:-}"
    if [ -z "$API_BASE_URL" ] && [ -z "$K8S_API_GATEWAY_URL" ] && [ -z "$K8S_INGRESS_HOST" ]; then
      echo "Set API_BASE_URL, K8S_API_GATEWAY_URL_PROD, or K8S_INGRESS_HOST_PROD."
      exit 1
    fi
    PRODUCTION_URL="${API_BASE_URL:-${K8S_API_GATEWAY_URL:-$K8S_INGRESS_HOST}}"
    if [[ "$PRODUCTION_URL" =~ ^http:// ]]; then
      echo "Production requires an HTTPS API Gateway URL."
      exit 1
    fi
    BUILD_NAME="1.0.0"
    BUILD_NUMBER="3"
    ;;
  *)
    echo "Unknown environment: $ENVIRONMENT"
    echo "Usage: $0 [dev|staging|prod] [android|windows|ios]"
    exit 1
    ;;
esac

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Flutter Build Configuration"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Environment:       $FLUTTER_ENV"
echo "Platform:          $PLATFORM"
echo "API Base URL:      $API_BASE_URL"
echo "K8s API Gateway:   $K8S_API_GATEWAY_URL"
echo "K8s Ingress Host:  $K8S_INGRESS_HOST"
echo "Build Name:        $BUILD_NAME"
echo "Build Number:      $BUILD_NUMBER"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Common dart defines
DART_DEFINES=(--dart-define="FLUTTER_ENV=$FLUTTER_ENV")

if [ -n "$API_BASE_URL" ]; then
  DART_DEFINES+=(--dart-define="API_BASE_URL=$API_BASE_URL")
fi

if [ -n "$K8S_API_GATEWAY_URL" ]; then
  DART_DEFINES+=(--dart-define="K8S_API_GATEWAY_URL=$K8S_API_GATEWAY_URL")
fi

if [ -n "$K8S_INGRESS_HOST" ]; then
  DART_DEFINES+=(--dart-define="K8S_INGRESS_HOST=$K8S_INGRESS_HOST")
fi

# Build based on platform
case "$PLATFORM" in
  android)
    echo "Building for Android..."
    flutter build apk \
      --build-name=$BUILD_NAME \
      --build-number=$BUILD_NUMBER \
      "${DART_DEFINES[@]}" \
      --release
    echo "✓ APK built: build/app/outputs/flutter-apk/app-release.apk"
    ;;
  windows)
    echo "Building for Windows..."
    flutter build windows \
      "${DART_DEFINES[@]}"
    echo "✓ Windows build completed: build/windows/runner/Release/"
    ;;
  ios)
    echo "Building for iOS..."
    flutter build ios \
      --build-name=$BUILD_NAME \
      --build-number=$BUILD_NUMBER \
      "${DART_DEFINES[@]}" \
      --release
    echo "✓ iOS build completed: build/ios/iphoneos/"
    ;;
  all)
    echo "Building for all platforms..."
    
    echo "1. Building for Android..."
    flutter build apk \
      --build-name=$BUILD_NAME \
      --build-number=$BUILD_NUMBER \
      "${DART_DEFINES[@]}" \
      --release
    
    echo "2. Building for Windows..."
    flutter build windows "${DART_DEFINES[@]}"
    
    echo "✓ All builds completed!"
    ;;
  *)
    echo "Unknown platform: $PLATFORM"
    echo "Usage: $0 [dev|staging|prod] [android|windows|ios|all]"
    exit 1
    ;;
esac

echo ""
echo "Build completed successfully!"
