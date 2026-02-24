---
applyTo: "lambdas/**/*.go"
---

# Go Lambda Code Instructions

## SDK & Runtime

- Use `aws-sdk-go-v2` for all AWS SDK interactions — never use `aws-sdk-go` (v1)
- Use `aws-lambda-go` for Lambda handler registration and event types
- Target `provided.al2023` runtime with `arm64` architecture

## Project Structure

- Define interfaces in `internal/port/` for all AWS service dependencies (S3, DynamoDB, etc.)
- Implement interfaces in `internal/storage/`
- Handler logic lives in `internal/handler/` — keep Lambda `main.go` files minimal (wire dependencies and call handler)
- Request/response types and DynamoDB item types go in `internal/model/`
- Validation rules (file type, size, PDF magic bytes) go in `internal/validation/`

## Handler Patterns

- Handler functions accept `context.Context` as the first parameter
- Return structured `events.APIGatewayProxyResponse` with appropriate HTTP status codes
- Use `slog` for structured logging — always include `fileId`, `userId`, and `status` fields
- Wrap errors with context: `fmt.Errorf("operation: %w", err)`

## Testing

- Unit tests use mock implementations of `port/` interfaces — no real AWS calls
- Test files live alongside the code they test (`_test.go` suffix)
- Use table-driven tests for validation logic
- Run tests: `cd lambdas && go test ./...`

## File Constraints

- PDF only (`application/pdf`), validated by magic bytes (`%PDF-`)
- Max file size: 1 MB (1,048,576 bytes)
- S3 key structure: `uploads/{userId}/{fileId}.pdf`
