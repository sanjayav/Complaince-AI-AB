#!/bin/bash

# AI CORE RAG API Service Deployment Script
# Phase 9: Deploy/Ops - Deployment Automation

set -e

# Configuration
SERVICE_NAME="ai-core-rag-api"
DOCKER_IMAGE="ai-core-rag-api:latest"
CONTAINER_NAME="ai-core-rag-api-container"
NETWORK_NAME="ai-core-network"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    log_info "Docker is running"
}

# Build the Docker image
build_image() {
    log_info "Building Docker image..."
    docker build -t $DOCKER_IMAGE .
    log_info "Docker image built successfully"
}

# Create Docker network if it doesn't exist
create_network() {
    if ! docker network ls | grep -q $NETWORK_NAME; then
        log_info "Creating Docker network: $NETWORK_NAME"
        docker network create $NETWORK_NAME
    else
        log_info "Docker network $NETWORK_NAME already exists"
    fi
}

# Stop and remove existing container
cleanup_container() {
    if docker ps -a | grep -q $CONTAINER_NAME; then
        log_info "Stopping existing container: $CONTAINER_NAME"
        docker stop $CONTAINER_NAME || true
        docker rm $CONTAINER_NAME || true
        log_info "Existing container cleaned up"
    fi
}

# Deploy the service
deploy_service() {
    log_info "Deploying $SERVICE_NAME..."
    
    # Run the container
    docker run -d \
        --name $CONTAINER_NAME \
        --network $NETWORK_NAME \
        -p 8000:8000 \
        --env-file .env \
        --restart unless-stopped \
        --health-cmd="curl -f http://localhost:8000/health || exit 1" \
        --health-interval=30s \
        --health-timeout=10s \
        --health-retries=3 \
        $DOCKER_IMAGE
    
    log_info "Service deployed successfully"
}

# Check service health
check_health() {
    log_info "Checking service health..."
    
    # Wait for service to start
    sleep 10
    
    # Check if container is running
    if ! docker ps | grep -q $CONTAINER_NAME; then
        log_error "Container is not running"
        docker logs $CONTAINER_NAME
        exit 1
    fi
    
    # Check health endpoint
    if curl -f http://localhost:8000/health > /dev/null 2>&1; then
        log_info "Service is healthy"
    else
        log_error "Service health check failed"
        docker logs $CONTAINER_NAME
        exit 1
    fi
}

# Show service status
show_status() {
    log_info "Service status:"
    docker ps --filter name=$CONTAINER_NAME
    
    log_info "Service logs (last 20 lines):"
    docker logs --tail 20 $CONTAINER_NAME
}

# Main deployment function
main() {
    log_info "Starting deployment of $SERVICE_NAME"
    
    check_docker
    build_image
    create_network
    cleanup_container
    deploy_service
    check_health
    show_status
    
    log_info "Deployment completed successfully!"
    log_info "Service is available at: http://localhost:8000"
    log_info "API documentation: http://localhost:8000/docs"
    log_info "Health check: http://localhost:8000/health"
}

# Handle command line arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "stop")
        log_info "Stopping service..."
        docker stop $CONTAINER_NAME || true
        log_info "Service stopped"
        ;;
    "start")
        log_info "Starting service..."
        docker start $CONTAINER_NAME || true
        log_info "Service started"
        ;;
    "restart")
        log_info "Restarting service..."
        docker restart $CONTAINER_NAME || true
        log_info "Service restarted"
        ;;
    "logs")
        docker logs -f $CONTAINER_NAME
        ;;
    "status")
        show_status
        ;;
    "cleanup")
        log_info "Cleaning up..."
        cleanup_container
        docker rmi $DOCKER_IMAGE || true
        log_info "Cleanup completed"
        ;;
    *)
        echo "Usage: $0 {deploy|stop|start|restart|logs|status|cleanup}"
        echo "  deploy  - Deploy the service (default)"
        echo "  stop    - Stop the service"
        echo "  start   - Start the service"
        echo "  restart - Restart the service"
        echo "  logs    - Show service logs"
        echo "  status  - Show service status"
        echo "  cleanup - Remove container and image"
        exit 1
        ;;
esac
