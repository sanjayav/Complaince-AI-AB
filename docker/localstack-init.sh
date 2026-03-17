#!/usr/bin/env bash
set -euo pipefail

awslocal() {
  aws --endpoint-url="${S3_ENDPOINT:-http://localhost:4566}" --region "${AWS_REGION:-us-east-1}" "$@"
}

BUCKET="${S3_BUCKET:-jlrdi}"

echo "Creating bucket ${BUCKET} on Localstack..."
awslocal s3api create-bucket --bucket "$BUCKET" >/dev/null 2>&1 || true
echo "Done."


