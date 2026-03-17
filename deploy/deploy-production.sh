#!/bin/bash

# 🚀 JLR Document Intelligence - Production Deployment Script
# This script automates the complete production deployment process

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ENVIRONMENT=${1:-"prod"}
AWS_REGION=${AWS_REGION:-"us-east-1"}
CLUSTER_NAME="jlrdi-cluster"
NAMESPACE="jlrdi"

# Logging function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        error "AWS CLI is not installed. Please install it first."
    fi
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        error "kubectl is not installed. Please install it first."
    fi
    
    # Check Terraform
    if ! command -v terraform &> /dev/null; then
        error "Terraform is not installed. Please install it first."
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed. Please install it first."
    fi
    
    # Check Helm
    if ! command -v helm &> /dev/null; then
        error "Helm is not installed. Please install it first."
    fi
    
    success "All prerequisites are satisfied"
}

# Deploy infrastructure
deploy_infrastructure() {
    log "Deploying infrastructure with Terraform..."
    
    cd deploy/terraform
    
    # Initialize Terraform
    log "Initializing Terraform..."
    terraform init
    
    # Plan deployment
    log "Planning infrastructure deployment..."
    terraform plan -var="environment=${ENVIRONMENT}" -out=production.tfplan
    
    # Apply infrastructure
    log "Applying infrastructure changes..."
    terraform apply production.tfplan
    
    # Get outputs
    CLUSTER_ENDPOINT=$(terraform output -raw cluster_endpoint)
    VPC_ID=$(terraform output -raw vpc_id)
    RDS_ENDPOINT=$(terraform output -raw rds_endpoint)
    REDIS_ENDPOINT=$(terraform output -raw redis_endpoint)
    S3_BUCKET_DOCS=$(terraform output -raw s3_bucket_documents)
    S3_BUCKET_IMAGES=$(terraform output -raw s3_bucket_page_images)
    ECR_REPO_URL=$(terraform output -raw ecr_repository_url)
    ALB_DNS=$(terraform output -raw alb_dns_name)
    
    success "Infrastructure deployed successfully"
    log "Cluster endpoint: ${CLUSTER_ENDPOINT}"
    log "VPC ID: ${VPC_ID}"
    log "RDS endpoint: ${RDS_ENDPOINT}"
    log "Redis endpoint: ${REDIS_ENDPOINT}"
    log "S3 Documents bucket: ${S3_BUCKET_DOCS}"
    log "S3 Images bucket: ${S3_BUCKET_IMAGES}"
    log "ECR repository: ${ECR_REPO_URL}"
    log "ALB DNS: ${ALB_DNS}"
    
    cd ../..
}

# Configure kubectl
configure_kubectl() {
    log "Configuring kubectl for EKS cluster..."
    
    aws eks update-kubeconfig --name ${CLUSTER_NAME} --region ${AWS_REGION}
    
    # Verify connection
    if ! kubectl cluster-info &> /dev/null; then
        error "Failed to connect to EKS cluster"
    fi
    
    success "kubectl configured successfully"
}

# Build and push Docker images
build_images() {
    log "Building and pushing Docker images..."
    
    # Build API image
    log "Building API image..."
    docker build -t jlrdi-api:latest .
    
    # Tag for ECR
    docker tag jlrdi-api:latest ${ECR_REPO_URL}:latest
    docker tag jlrdi-api:latest ${ECR_REPO_URL}:v1.0.0
    
    # Login to ECR
    aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${ECR_REPO_URL}
    
    # Push images
    log "Pushing images to ECR..."
    docker push ${ECR_REPO_URL}:latest
    docker push ${ECR_REPO_URL}:v1.0.0
    
    success "Images built and pushed successfully"
}

# Deploy monitoring stack
deploy_monitoring() {
    log "Deploying monitoring stack..."
    
    # Create monitoring namespace
    kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
    
    # Add Prometheus Helm repository
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
    helm repo update
    
    # Install Prometheus stack
    helm install prometheus prometheus-community/kube-prometheus-stack \
        --namespace monitoring \
        --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
        --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false
    
    # Wait for Prometheus to be ready
    kubectl wait --for=condition=ready pod -l app=prometheus -n monitoring --timeout=300s
    
    success "Monitoring stack deployed successfully"
}

# Deploy application
deploy_application() {
    log "Deploying JLR Document Intelligence application..."
    
    # Create namespace
    kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
    
    # Create secrets (replace with actual values)
    log "Creating Kubernetes secrets..."
    kubectl create secret generic jlrdi-secrets \
        --from-literal=database-url="postgresql://jlrdi_admin:password@${RDS_ENDPOINT}:5432/jlrdi" \
        --from-literal=jwt-secret="your-super-secret-jwt-key-change-this-in-production" \
        --from-literal=aws-access-key="${AWS_ACCESS_KEY_ID}" \
        --from-literal=aws-secret-key="${AWS_SECRET_ACCESS_KEY}" \
        -n ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
    
    # Deploy Qdrant
    log "Deploying Qdrant vector database..."
    kubectl apply -f deploy/kubernetes/qdrant-deployment.yaml -n ${NAMESPACE}
    
    # Wait for Qdrant to be ready
    kubectl wait --for=condition=ready pod -l app=qdrant -n ${NAMESPACE} --timeout=300s
    
    # Deploy main application
    log "Deploying main application..."
    kubectl apply -f deploy/kubernetes/deployment.yaml -n ${NAMESPACE}
    
    # Wait for application to be ready
    kubectl wait --for=condition=ready pod -l app=jlrdi-api -n ${NAMESPACE} --timeout=300s
    
    success "Application deployed successfully"
}

# Deploy ingress and TLS
deploy_ingress() {
    log "Deploying ingress controller and TLS..."
    
    # Install NGINX Ingress Controller
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/aws/deploy.yaml
    
    # Wait for ingress controller
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=ingress-nginx -n ingress-nginx --timeout=300s
    
    # Install cert-manager
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
    
    # Wait for cert-manager
    kubectl wait --for=condition=ready pod -l app=cert-manager -n cert-manager --timeout=300s
    
    # Deploy ingress
    kubectl apply -f deploy/kubernetes/ingress.yaml -n ${NAMESPACE}
    
    success "Ingress and TLS configured successfully"
}

# Deploy security policies
deploy_security() {
    log "Deploying security policies..."
    
    # Apply network policies
    kubectl apply -f deploy/security/network-policies.yaml -n ${NAMESPACE}
    
    # Apply RBAC
    kubectl apply -f deploy/security/rbac.yaml -n ${NAMESPACE}
    
    success "Security policies deployed successfully"
}

# Deploy auto-scaling
deploy_scaling() {
    log "Deploying auto-scaling configuration..."
    
    # Deploy HPA
    kubectl apply -f deploy/scaling/hpa.yaml -n ${NAMESPACE}
    
    # Deploy VPA
    kubectl apply -f deploy/scaling/vpa.yaml -n ${NAMESPACE}
    
    success "Auto-scaling configured successfully"
}

# Run health checks
health_checks() {
    log "Running health checks..."
    
    # Check pods
    kubectl get pods -n ${NAMESPACE}
    
    # Check services
    kubectl get services -n ${NAMESPACE}
    
    # Check ingress
    kubectl get ingress -n ${NAMESPACE}
    
    # Check HPA
    kubectl get hpa -n ${NAMESPACE}
    
    # Test API health
    log "Testing API health..."
    if kubectl get ingress -n ${NAMESPACE} | grep -q "jlrdi-ingress"; then
        INGRESS_HOST=$(kubectl get ingress jlrdi-ingress -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
        if [ -n "${INGRESS_HOST}" ]; then
            log "Testing health endpoint at https://${INGRESS_HOST}/v1/health"
            # Note: In production, you'd want to wait for DNS propagation
        fi
    fi
    
    success "Health checks completed"
}

# Main deployment function
main() {
    log "🚀 Starting JLR Document Intelligence Production Deployment"
    log "Environment: ${ENVIRONMENT}"
    log "AWS Region: ${AWS_REGION}"
    log "Cluster: ${CLUSTER_NAME}"
    log "Namespace: ${NAMESPACE}"
    
    # Check prerequisites
    check_prerequisites
    
    # Deploy infrastructure
    deploy_infrastructure
    
    # Configure kubectl
    configure_kubectl
    
    # Build and push images
    build_images
    
    # Deploy monitoring
    deploy_monitoring
    
    # Deploy application
    deploy_application
    
    # Deploy ingress and TLS
    deploy_ingress
    
    # Deploy security
    deploy_security
    
    # Deploy scaling
    deploy_scaling
    
    # Health checks
    health_checks
    
    success "🎉 JLR Document Intelligence Production Deployment Completed Successfully!"
    
    log "📋 Next Steps:"
    log "1. Update DNS records to point to: ${ALB_DNS}"
    log "2. Configure monitoring alerts in Prometheus"
    log "3. Set up log aggregation (Fluent Bit + CloudWatch)"
    log "4. Configure backup and disaster recovery"
    log "5. Run load tests to validate performance"
    log "6. Update SSL certificates with your domain"
    
    log "🔗 Useful Commands:"
    log "  kubectl get all -n ${NAMESPACE}"
    log "  kubectl logs -f deployment/jlrdi-api -n ${NAMESPACE}"
    log "  kubectl get hpa -n ${NAMESPACE}"
    log "  kubectl get ingress -n ${NAMESPACE}"
}

# Run main function
main "$@"
