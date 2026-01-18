#!/usr/bin/env bash
# AWS LocalStack Test Suite for kspec
# Tests AWS security policies against LocalStack
#
# Usage: ./scripts/test-aws-localstack.sh [--setup-only] [--test-only] [--cleanup]
#
# Requirements:
# - Docker
# - AWS CLI v2
# - Go (for building kspec)

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
LOCALSTACK_CONTAINER="kspec-localstack-test"
LOCALSTACK_PORT=4566
LOCALSTACK_IMAGE="localstack/localstack:latest"
AWS_REGION="us-east-1"
AWS_ACCESS_KEY_ID="test"
AWS_SECRET_ACCESS_KEY="test"
ENDPOINT_URL="http://localhost:${LOCALSTACK_PORT}"

# Test data
TEST_ORG="kspec-test-org"
TEST_BUCKET_SECURE="kspec-secure-bucket"
TEST_BUCKET_INSECURE="kspec-insecure-bucket"
TEST_USER_WITH_MFA="test-user-mfa"
TEST_USER_WITHOUT_MFA="test-user-no-mfa"
TEST_ROLE="kspec-test-role"
TEST_GROUP="kspec-test-group"
TEST_KMS_KEY_ALIAS="alias/kspec-test-key"
TEST_SECRET="kspec-test-secret"
TEST_QUEUE="kspec-test-queue"
TEST_TOPIC="kspec-test-topic"
TEST_TABLE="kspec-test-table"
TEST_LOG_GROUP="/kspec/test"

# Counters
RESOURCES_CREATED=0
RESOURCES_FAILED=0

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_section() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Strip ANSI escape codes (using perl for cross-platform compatibility)
strip_ansi() {
    perl -pe 's/\e\[[0-9;]*m//g'
}

# AWS CLI wrapper with endpoint URL
aws_local() {
    aws --endpoint-url="${ENDPOINT_URL}" --region "${AWS_REGION}" "$@"
}

# Check prerequisites
check_prerequisites() {
    log_section "Checking Prerequisites"

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    log_success "Docker installed"

    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed"
        exit 1
    fi
    log_success "AWS CLI installed"

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    log_success "Go installed"
}

# Start LocalStack
start_localstack() {
    log_section "Starting LocalStack"

    # Stop existing container if running
    if docker ps -q -f name="${LOCALSTACK_CONTAINER}" | grep -q .; then
        log_info "Stopping existing LocalStack container..."
        docker stop "${LOCALSTACK_CONTAINER}" > /dev/null 2>&1 || true
        docker rm "${LOCALSTACK_CONTAINER}" > /dev/null 2>&1 || true
    fi

    # Start LocalStack with all services we need
    log_info "Starting LocalStack container..."
    docker run -d \
        --name "${LOCALSTACK_CONTAINER}" \
        -p "${LOCALSTACK_PORT}:4566" \
        -e SERVICES=s3,iam,sts,kms,secretsmanager,sqs,sns,dynamodb,logs,ssm,ec2,lambda \
        -e DOCKER_HOST=unix:///var/run/docker.sock \
        -v /var/run/docker.sock:/var/run/docker.sock \
        "${LOCALSTACK_IMAGE}" > /dev/null

    # Wait for LocalStack to be ready
    log_info "Waiting for LocalStack to be ready..."
    local retries=30
    while ! curl -s "${ENDPOINT_URL}/_localstack/health" | grep -q '"s3": "available"'; do
        retries=$((retries - 1))
        if [ $retries -eq 0 ]; then
            log_error "LocalStack failed to start"
            docker logs "${LOCALSTACK_CONTAINER}"
            exit 1
        fi
        sleep 2
    done

    log_success "LocalStack is ready"
}

# Create test resource with error handling
create_resource() {
    local name="$1"
    local cmd="$2"

    if eval "$cmd" > /dev/null 2>&1; then
        log_success "Created: $name"
        RESOURCES_CREATED=$((RESOURCES_CREATED + 1))
        return 0
    else
        log_warn "Failed to create: $name"
        RESOURCES_FAILED=$((RESOURCES_FAILED + 1))
        return 1
    fi
}

# Setup IAM resources
setup_iam() {
    log_section "Setting up IAM Resources"

    # Create IAM users
    create_resource "IAM User (with MFA simulation)" \
        "aws_local iam create-user --user-name ${TEST_USER_WITH_MFA}"

    create_resource "IAM User (without MFA)" \
        "aws_local iam create-user --user-name ${TEST_USER_WITHOUT_MFA}"

    # Create access keys for users (to test rotation policies)
    create_resource "Access Key for ${TEST_USER_WITH_MFA}" \
        "aws_local iam create-access-key --user-name ${TEST_USER_WITH_MFA}"

    # Create IAM role
    local trust_policy='{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"},"Action":"sts:AssumeRole"}]}'
    create_resource "IAM Role" \
        "aws_local iam create-role --role-name ${TEST_ROLE} --assume-role-policy-document '${trust_policy}'"

    # Create IAM group
    create_resource "IAM Group" \
        "aws_local iam create-group --group-name ${TEST_GROUP}"

    # Add user to group
    create_resource "Add user to group" \
        "aws_local iam add-user-to-group --user-name ${TEST_USER_WITH_MFA} --group-name ${TEST_GROUP}"

    # Create inline policy (to test no-inline-policies check)
    local inline_policy='{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}'
    create_resource "Inline Policy (for testing)" \
        "aws_local iam put-user-policy --user-name ${TEST_USER_WITHOUT_MFA} --policy-name InlineTestPolicy --policy-document '${inline_policy}'"

    # Create managed policy with admin privileges (to test no-admin-privileges check)
    local admin_policy='{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}'
    create_resource "Admin Policy (for testing)" \
        "aws_local iam create-policy --policy-name AdminTestPolicy --policy-document '${admin_policy}'"

    # Create restricted policy (compliant)
    local restricted_policy='{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject","s3:ListBucket"],"Resource":["arn:aws:s3:::specific-bucket","arn:aws:s3:::specific-bucket/*"]}]}'
    create_resource "Restricted Policy (compliant)" \
        "aws_local iam create-policy --policy-name RestrictedTestPolicy --policy-document '${restricted_policy}'"

    # Set password policy
    create_resource "Password Policy" \
        "aws_local iam update-account-password-policy \
            --minimum-password-length 14 \
            --require-symbols \
            --require-numbers \
            --require-uppercase-characters \
            --require-lowercase-characters \
            --max-password-age 90 \
            --password-reuse-prevention 24"
}

# Setup S3 resources
setup_s3() {
    log_section "Setting up S3 Resources"

    # Create secure bucket with encryption and versioning
    create_resource "S3 Bucket (secure)" \
        "aws_local s3 mb s3://${TEST_BUCKET_SECURE}"

    # Enable versioning
    create_resource "S3 Versioning (secure bucket)" \
        "aws_local s3api put-bucket-versioning --bucket ${TEST_BUCKET_SECURE} --versioning-configuration Status=Enabled"

    # Enable encryption
    create_resource "S3 Encryption (secure bucket)" \
        "aws_local s3api put-bucket-encryption --bucket ${TEST_BUCKET_SECURE} --server-side-encryption-configuration '{\"Rules\":[{\"ApplyServerSideEncryptionByDefault\":{\"SSEAlgorithm\":\"AES256\"}}]}'"

    # Block public access
    create_resource "S3 Public Access Block (secure bucket)" \
        "aws_local s3api put-public-access-block --bucket ${TEST_BUCKET_SECURE} --public-access-block-configuration BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

    # Create insecure bucket (for testing failures)
    create_resource "S3 Bucket (insecure - for testing)" \
        "aws_local s3 mb s3://${TEST_BUCKET_INSECURE}"

    # Note: Intentionally NOT enabling encryption, versioning, or public access block
    log_info "Insecure bucket created without security features (for testing policy violations)"
}

# Setup KMS resources
setup_kms() {
    log_section "Setting up KMS Resources"

    # Create KMS key with rotation
    local key_id
    key_id=$(aws_local kms create-key --description "kspec test key" --query 'KeyMetadata.KeyId' --output text 2>/dev/null || echo "")

    if [ -n "$key_id" ]; then
        log_success "Created: KMS Key ($key_id)"
        RESOURCES_CREATED=$((RESOURCES_CREATED + 1))

        # Enable key rotation
        create_resource "KMS Key Rotation" \
            "aws_local kms enable-key-rotation --key-id ${key_id}"

        # Create alias
        create_resource "KMS Key Alias" \
            "aws_local kms create-alias --alias-name ${TEST_KMS_KEY_ALIAS} --target-key-id ${key_id}"
    else
        log_warn "Failed to create: KMS Key"
        RESOURCES_FAILED=$((RESOURCES_FAILED + 1))
    fi
}

# Setup Secrets Manager resources
setup_secrets() {
    log_section "Setting up Secrets Manager Resources"

    # Create secret with rotation (simulated)
    create_resource "Secret" \
        "aws_local secretsmanager create-secret --name ${TEST_SECRET} --secret-string '{\"username\":\"admin\",\"password\":\"test123\"}'"

    # Create secret without rotation (for testing)
    create_resource "Secret (no rotation - for testing)" \
        "aws_local secretsmanager create-secret --name ${TEST_SECRET}-no-rotation --secret-string 'simple-secret'"
}

# Setup SQS resources
setup_sqs() {
    log_section "Setting up SQS Resources"

    # Create encrypted queue
    create_resource "SQS Queue (encrypted)" \
        "aws_local sqs create-queue --queue-name ${TEST_QUEUE}-secure --attributes KmsMasterKeyId=alias/aws/sqs"

    # Create queue without encryption (for testing)
    create_resource "SQS Queue (no encryption - for testing)" \
        "aws_local sqs create-queue --queue-name ${TEST_QUEUE}-insecure"

    # Create dead letter queue
    local dlq_url
    dlq_url=$(aws_local sqs create-queue --queue-name ${TEST_QUEUE}-dlq --query 'QueueUrl' --output text 2>/dev/null || echo "")
    if [ -n "$dlq_url" ]; then
        log_success "Created: SQS DLQ"
        RESOURCES_CREATED=$((RESOURCES_CREATED + 1))

        # Get DLQ ARN and configure redrive policy
        local dlq_arn
        dlq_arn=$(aws_local sqs get-queue-attributes --queue-url "$dlq_url" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text 2>/dev/null || echo "")
        if [ -n "$dlq_arn" ]; then
            create_resource "SQS Redrive Policy" \
                "aws_local sqs set-queue-attributes --queue-url $(aws_local sqs get-queue-url --queue-name ${TEST_QUEUE}-secure --query 'QueueUrl' --output text) --attributes '{\"RedrivePolicy\":\"{\\\"deadLetterTargetArn\\\":\\\"${dlq_arn}\\\",\\\"maxReceiveCount\\\":\\\"5\\\"}\"}'"
        fi
    fi
}

# Setup SNS resources
setup_sns() {
    log_section "Setting up SNS Resources"

    # Create encrypted topic
    create_resource "SNS Topic (encrypted)" \
        "aws_local sns create-topic --name ${TEST_TOPIC}-secure --attributes KmsMasterKeyId=alias/aws/sns"

    # Create topic without encryption (for testing)
    create_resource "SNS Topic (no encryption - for testing)" \
        "aws_local sns create-topic --name ${TEST_TOPIC}-insecure"
}

# Setup DynamoDB resources
setup_dynamodb() {
    log_section "Setting up DynamoDB Resources"

    # Create table with encryption and PITR
    create_resource "DynamoDB Table" \
        "aws_local dynamodb create-table \
            --table-name ${TEST_TABLE} \
            --attribute-definitions AttributeName=id,AttributeType=S \
            --key-schema AttributeName=id,KeyType=HASH \
            --billing-mode PAY_PER_REQUEST"

    # Wait for table to be active
    sleep 2

    # Enable PITR
    create_resource "DynamoDB PITR" \
        "aws_local dynamodb update-continuous-backups \
            --table-name ${TEST_TABLE} \
            --point-in-time-recovery-specification PointInTimeRecoveryEnabled=true"
}

# Setup CloudWatch resources
setup_cloudwatch() {
    log_section "Setting up CloudWatch Resources"

    # Create log group with retention
    create_resource "CloudWatch Log Group" \
        "aws_local logs create-log-group --log-group-name ${TEST_LOG_GROUP}"

    create_resource "CloudWatch Log Group Retention" \
        "aws_local logs put-retention-policy --log-group-name ${TEST_LOG_GROUP} --retention-in-days 30"

    # Create log group without retention (for testing)
    create_resource "CloudWatch Log Group (no retention - for testing)" \
        "aws_local logs create-log-group --log-group-name ${TEST_LOG_GROUP}-no-retention"
}

# Setup SSM resources
setup_ssm() {
    log_section "Setting up SSM Resources"

    # Create SecureString parameter
    create_resource "SSM SecureString Parameter" \
        "aws_local ssm put-parameter \
            --name /kspec/test/secure \
            --value 'secret-value' \
            --type SecureString"

    # Create String parameter (for comparison)
    create_resource "SSM String Parameter" \
        "aws_local ssm put-parameter \
            --name /kspec/test/plain \
            --value 'plain-value' \
            --type String"
}

# Setup EC2 resources (limited in LocalStack)
setup_ec2() {
    log_section "Setting up EC2 Resources"

    # Create VPC
    local vpc_id
    vpc_id=$(aws_local ec2 create-vpc --cidr-block 10.0.0.0/16 --query 'Vpc.VpcId' --output text 2>/dev/null || echo "")

    if [ -n "$vpc_id" ]; then
        log_success "Created: VPC ($vpc_id)"
        RESOURCES_CREATED=$((RESOURCES_CREATED + 1))

        # Tag VPC
        aws_local ec2 create-tags --resources "$vpc_id" --tags Key=Name,Value=kspec-test-vpc > /dev/null 2>&1 || true

        # Create security group with restricted rules (compliant)
        local sg_secure_id
        sg_secure_id=$(aws_local ec2 create-security-group \
            --group-name kspec-sg-secure \
            --description "Secure SG for kspec testing" \
            --vpc-id "$vpc_id" \
            --query 'GroupId' --output text 2>/dev/null || echo "")

        if [ -n "$sg_secure_id" ]; then
            log_success "Created: Security Group (secure)"
            RESOURCES_CREATED=$((RESOURCES_CREATED + 1))

            # Add restricted SSH rule (specific IP)
            create_resource "SG Rule (restricted SSH)" \
                "aws_local ec2 authorize-security-group-ingress \
                    --group-id $sg_secure_id \
                    --protocol tcp \
                    --port 22 \
                    --cidr 10.0.0.0/8"
        fi

        # Create security group with open rules (non-compliant)
        local sg_insecure_id
        sg_insecure_id=$(aws_local ec2 create-security-group \
            --group-name kspec-sg-insecure \
            --description "Insecure SG for kspec testing" \
            --vpc-id "$vpc_id" \
            --query 'GroupId' --output text 2>/dev/null || echo "")

        if [ -n "$sg_insecure_id" ]; then
            log_success "Created: Security Group (insecure - for testing)"
            RESOURCES_CREATED=$((RESOURCES_CREATED + 1))

            # Add open SSH rule (0.0.0.0/0) - non-compliant
            create_resource "SG Rule (open SSH - for testing)" \
                "aws_local ec2 authorize-security-group-ingress \
                    --group-id $sg_insecure_id \
                    --protocol tcp \
                    --port 22 \
                    --cidr 0.0.0.0/0"

            # Add open RDP rule - non-compliant
            create_resource "SG Rule (open RDP - for testing)" \
                "aws_local ec2 authorize-security-group-ingress \
                    --group-id $sg_insecure_id \
                    --protocol tcp \
                    --port 3389 \
                    --cidr 0.0.0.0/0"
        fi
    else
        log_warn "Failed to create: VPC"
        RESOURCES_FAILED=$((RESOURCES_FAILED + 1))
    fi
}

# Build kspec
build_kspec() {
    log_section "Building kspec"

    log_info "Building kspec..."
    if go build -o /tmp/kspec ./cmd/kspec; then
        log_success "kspec built successfully"
    else
        log_error "Failed to build kspec"
        exit 1
    fi
}

# Run kspec scan
run_scan() {
    log_section "Running kspec Scan"

    export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}"
    export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}"
    export AWS_DEFAULT_REGION="${AWS_REGION}"

    log_info "Running discovery..."
    /tmp/kspec discover aws --endpoint-url="${ENDPOINT_URL}" --sequential

    echo ""
    log_info "Running policy scan..."
    /tmp/kspec scan aws --endpoint-url="${ENDPOINT_URL}" --no-ui 2>&1 | tee /tmp/kspec-scan-output.log

    echo ""
    log_info "Generating HTML report..."
    /tmp/kspec scan aws --endpoint-url="${ENDPOINT_URL}" --no-ui --export /tmp/kspec-report.html 2>/dev/null

    if [ -f /tmp/kspec-report.html ]; then
        log_success "HTML report generated: /tmp/kspec-report.html"
        local report_size
        report_size=$(wc -c < /tmp/kspec-report.html | tr -d ' ')
        log_info "Report size: ${report_size} bytes"

        # Verify HTML structure
        if grep -q "<html" /tmp/kspec-report.html && grep -q "</html>" /tmp/kspec-report.html; then
            log_success "HTML report structure is valid"
        else
            log_warn "HTML report may be malformed"
        fi

        # Count check results in HTML
        local html_passed html_failed
        html_passed=$(grep -c "check-pass\|status-pass\|✓\|PASS" /tmp/kspec-report.html 2>/dev/null || echo "0")
        html_failed=$(grep -c "check-fail\|status-fail\|✗\|FAIL" /tmp/kspec-report.html 2>/dev/null || echo "0")
        log_info "HTML report contains ~${html_passed} pass indicators and ~${html_failed} fail indicators"
    else
        log_warn "HTML report was not generated"
    fi

    echo ""
    log_section "Scan Results Analysis"

    # Extract summary from kspec output (e.g., "failed=25 passed=12")
    # Strip ANSI escape codes first for reliable parsing (using perl for cross-platform compat)
    local summary_line
    summary_line=$(grep "Scan summary" /tmp/kspec-scan-output.log 2>/dev/null | perl -pe 's/\e\[[0-9;]*m//g' || echo "")

    if [ -n "$summary_line" ]; then
        # Parse from kspec's structured output (compatible with macOS and Linux)
        local passed failed skipped total
        passed=$(echo "$summary_line" | sed -n 's/.*passed=\([0-9]*\).*/\1/p')
        failed=$(echo "$summary_line" | sed -n 's/.*failed=\([0-9]*\).*/\1/p')
        skipped=$(echo "$summary_line" | sed -n 's/.*skipped=\([0-9]*\).*/\1/p')
        total=$(echo "$summary_line" | sed -n 's/.*total=\([0-9]*\).*/\1/p')

        echo ""
        echo -e "${GREEN}Passed:${NC}  ${passed:-0}"
        echo -e "${RED}Failed:${NC}  ${failed:-0}"
        echo -e "${YELLOW}Skipped:${NC} ${skipped:-0}"
        echo -e "${BLUE}Total:${NC}   ${total:-0}"
    else
        # Fallback: count from log output
        local passed failed
        passed=$(grep -c "status=PASS" /tmp/kspec-scan-output.log 2>/dev/null || echo "0")
        failed=$(grep -c "status=FAIL" /tmp/kspec-scan-output.log 2>/dev/null || echo "0")

        echo ""
        echo -e "${GREEN}Passed:${NC} $passed"
        echo -e "${RED}Failed:${NC} $failed"
    fi

    # Show unique failed checks (strip ANSI first, then grep)
    echo ""
    log_info "Unique failed checks:"
    local failed_checks
    failed_checks=$(cat /tmp/kspec-scan-output.log 2>/dev/null | strip_ansi | grep "status=FAIL" | \
        sed 's/.*id=//' | sed 's/ .*//' | sort -u)
    if [ -n "$failed_checks" ]; then
        echo "$failed_checks" | while read -r check_id; do
            echo -e "  ${RED}✗${NC} $check_id"
        done
    fi

    # Show unique passed checks (strip ANSI first, then grep)
    echo ""
    log_info "Unique passed checks:"
    local passed_checks
    passed_checks=$(cat /tmp/kspec-scan-output.log 2>/dev/null | strip_ansi | grep "status=PASS" | \
        sed 's/.*id=//' | sed 's/ .*//' | sort -u)
    if [ -n "$passed_checks" ]; then
        echo "$passed_checks" | while read -r check_id; do
            echo -e "  ${GREEN}✓${NC} $check_id"
        done
    fi
}

# Cleanup
cleanup() {
    log_section "Cleanup"

    log_info "Stopping LocalStack container..."
    docker stop "${LOCALSTACK_CONTAINER}" > /dev/null 2>&1 || true
    docker rm "${LOCALSTACK_CONTAINER}" > /dev/null 2>&1 || true

    log_info "Removing temporary files..."
    rm -f /tmp/kspec /tmp/kspec-scan-output.log /tmp/kspec-report.html

    log_success "Cleanup complete"
}

# Print summary
print_summary() {
    log_section "Test Setup Summary"

    echo ""
    echo -e "Resources created: ${GREEN}${RESOURCES_CREATED}${NC}"
    echo -e "Resources failed:  ${YELLOW}${RESOURCES_FAILED}${NC}"
    echo ""
    echo "LocalStack endpoint: ${ENDPOINT_URL}"
    echo ""
}

# Main execution
main() {
    local setup_only=false
    local test_only=false
    local cleanup_only=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --setup-only)
                setup_only=true
                shift
                ;;
            --test-only)
                test_only=true
                shift
                ;;
            --cleanup)
                cleanup_only=true
                shift
                ;;
            *)
                echo "Unknown option: $1"
                echo "Usage: $0 [--setup-only] [--test-only] [--cleanup]"
                exit 1
                ;;
        esac
    done

    # Cleanup only
    if [ "$cleanup_only" = true ]; then
        cleanup
        exit 0
    fi

    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║           kspec AWS LocalStack Test Suite                     ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    # Test only (assumes LocalStack is running)
    if [ "$test_only" = true ]; then
        build_kspec
        run_scan
        exit 0
    fi

    # Full setup or setup-only
    check_prerequisites
    start_localstack

    # Setup all resources
    setup_iam
    setup_s3
    setup_kms
    setup_secrets
    setup_sqs
    setup_sns
    setup_dynamodb
    setup_cloudwatch
    setup_ssm
    setup_ec2

    print_summary

    if [ "$setup_only" = true ]; then
        log_info "Setup complete. Run with --test-only to execute scan."
        log_info "Run with --cleanup to remove LocalStack container."
        exit 0
    fi

    # Build and run scan
    build_kspec
    run_scan

    echo ""
    log_info "Test complete. Run with --cleanup to remove LocalStack container."
}

# Run main
main "$@"
