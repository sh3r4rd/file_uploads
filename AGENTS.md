# AGENTS.md

Universal agent instructions for AI coding tools (Copilot, Codex, Jules, Cursor, Claude Code, and others).

## Project Overview

Serverless file upload service on AWS using presigned URLs for secure PDF uploads. Two-step upload flow: client requests a presigned URL via API Gateway/Lambda, then PUTs the file directly to S3. An S3 event triggers a second Lambda to validate and confirm the upload.

## Build & Development Commands

```bash
make build          # Cross-compile Go lambdas for Amazon Linux 2023 ARM64
make test           # Run all Go tests (cd lambdas && go test ./...)
make lint           # Run go vet on lambdas
make fmt            # Format Go code + Terraform files
make init           # Initialize Terraform (cd infra && terraform init)
make validate       # Validate Terraform config
make deploy         # Build + deploy via Terraform
make clean          # Remove compiled binaries (lambdas/bin/)
```

Run a single test:
```bash
cd lambdas && go test ./internal/validation/ -run TestValidateName
```

## Architecture

### Request Flow

1. `POST /uploads` → API Gateway (API key auth) → `RequestUploadHandler` Lambda → validates input, writes PENDING record to DynamoDB, returns presigned S3 PUT URL
2. Client PUTs PDF directly to S3 using presigned URL
3. S3 event notification → `ConfirmUploadHandler` Lambda → validates PDF (magic bytes, size), updates DynamoDB status to UPLOADED or REJECTED

### Lambda Functions

Two Lambda functions written in Go (`provided.al2023` runtime, ARM64):

- `lambdas/cmd/request-upload/` — API Gateway trigger, generates presigned URLs
- `lambdas/cmd/confirm-upload/` — S3 event trigger, validates uploaded files

### Shared Go Packages

Located in `lambdas/internal/`:

| Package | Purpose |
|---------|---------|
| `handler/` | Core handler logic for each Lambda |
| `validation/` | File validation rules (type, size, PDF magic bytes) |
| `storage/` | S3 and DynamoDB operations |
| `port/` | Interfaces for external dependencies (enables mocking in tests) |
| `model/` | Request/response structs, DynamoDB item types |

### Infrastructure

Terraform HCL in `infra/`, split by resource type: `s3.tf`, `dynamodb.tf`, `sqs.tf`, `iam.tf`, `lambda.tf`, `apigateway.tf`. Each Lambda gets its own IAM role with least-privilege permissions scoped to specific resource ARNs.

## Key Technical Details

- Go module: `github.com/sh3r4rd/file_uploads` (in `lambdas/`)
- Uses `aws-sdk-go-v2` (not v1) and `aws-lambda-go`
- S3 key structure: `uploads/{userId}/{fileId}.pdf`
- DynamoDB table: `FileUploads` with `fileId` (PK) and GSI on `userId-createdAt`
- Presigned URLs have 5-minute TTL with content-type and content-length conditions
- File constraints: PDF only (`application/pdf`), max 1 MB (1,048,576 bytes)
- Terraform variables configured via `infra/terraform.tfvars` (copy from `terraform.tfvars.example`)

## Code Style

### Go Conventions

- Use `aws-sdk-go-v2` for all AWS SDK interactions — never `aws-sdk-go` (v1)
- Define interfaces in `internal/port/` for every external AWS service dependency
- Handler functions accept `context.Context` as the first parameter
- Return structured API Gateway proxy responses (`events.APIGatewayProxyResponse`) with appropriate HTTP status codes
- Use `slog` for structured logging; include `fileId`, `userId`, and `status` fields
- Unit tests use mock implementations of `port/` interfaces — no real AWS calls in tests
- Error messages should be lowercase and descriptive

### Terraform Conventions

- One `.tf` file per AWS service (e.g., `s3.tf`, `dynamodb.tf`, `lambda.tf`)
- Apply `local.common_tags` to all taggable resources
- Use variables for configurable values; define them in `variables.tf`
- AWS provider version constraint: `~> 5.0`
- Terraform version: `>= 1.5`
- Use `provided.al2023` runtime with `arm64` architecture for all Lambda functions

## Prerequisites

- Go 1.25+ (go.mod specifies 1.25.1)
- Terraform >= 1.5
- AWS CLI configured with appropriate credentials

## Git Workflow

- Create feature branches from `main`
- Write clear, conventional commit messages (e.g., `feat:`, `fix:`, `docs:`, `refactor:`)
- Run `make test` and `make lint` before committing
- Keep PRs focused on a single concern

## Boundaries

### Always

- Run `make test` after any Go code change
- Run `make validate` after any Terraform change
- Use interfaces from `internal/port/` for AWS service dependencies
- Scope IAM permissions to specific resource ARNs (least privilege)

### Never

- Commit AWS credentials, secrets, or `.tfvars` files with real values
- Use `aws-sdk-go` v1 — always use `aws-sdk-go-v2`
- Bypass presigned URL validation or remove content-type/size constraints
- Add `*` resource ARNs to IAM policies
