#!/bin/bash
set -e

# Usage information
function show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Build and push Docker images to Google Container Registry or Artifact Registry"
    echo ""
    echo "Options:"
    echo "  -p, --project-id PROJECT_ID    Google Cloud Project ID (required)"
    echo "  -r, --registry REGISTRY        Container registry location (ex: asia-northeast1-docker.pkg.dev/PROJECT_ID/REPO_NAME) (required)"
    echo "  -t, --tag TAG                  Image tag (default: latest)"
    echo "  -f, --frontend-only            Build and push frontend only"
    echo "  -b, --backend-only             Build and push backend only"
    echo "  -h, --help                     Show this help text"
    exit 1
}

# Default values
TAG="latest"
BUILD_FRONTEND=true
BUILD_BACKEND=true

# Parse arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -p|--project-id)
            PROJECT_ID="$2"
            shift 2
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -f|--frontend-only)
            BUILD_FRONTEND=true
            BUILD_BACKEND=false
            shift
            ;;
        -b|--backend-only)
            BUILD_FRONTEND=false
            BUILD_BACKEND=true
            shift
            ;;
        -h|--help)
            show_usage
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            ;;
    esac
done

# Check required arguments
if [ -z "$PROJECT_ID" ] || [ -z "$REGISTRY" ]; then
    echo "Error: PROJECT_ID and REGISTRY are required"
    show_usage
fi

# Set working directory to repository root
cd "$(dirname "$0")/../../../"

# Authenticate with Google Cloud (if needed)
gcloud auth configure-docker

# Build and push frontend
if [ "$BUILD_FRONTEND" = true ]; then
    echo "Building frontend image..."
    FRONTEND_IMAGE="${REGISTRY}/golink-frontend:${TAG}"
    docker build -t "$FRONTEND_IMAGE" -f frontend/Dockerfile ./frontend

    echo "Pushing frontend image to $FRONTEND_IMAGE..."
    docker push "$FRONTEND_IMAGE"

    echo "Frontend image built and pushed successfully."
    echo "Frontend image: $FRONTEND_IMAGE"
fi

# Build and push backend
if [ "$BUILD_BACKEND" = true ]; then
    echo "Building backend image..."
    BACKEND_IMAGE="${REGISTRY}/golink-backend:${TAG}"
    docker build -t "$BACKEND_IMAGE" -f backend/Dockerfile ./backend

    echo "Pushing backend image to $BACKEND_IMAGE..."
    docker push "$BACKEND_IMAGE"

    echo "Backend image built and pushed successfully."
    echo "Backend image: $BACKEND_IMAGE"
fi

echo "Done!"
