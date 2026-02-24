# Copilot Instructions

Serverless file upload service on AWS — Go Lambdas with Terraform infrastructure. See [AGENTS.md](../AGENTS.md) for full project context, architecture, and coding standards.

## Code Review Preferences

- Flag any use of `aws-sdk-go` v1 — this project uses `aws-sdk-go-v2` exclusively
- Verify IAM policies use specific resource ARNs, not wildcards
- Check that new AWS service dependencies have a corresponding interface in `lambdas/internal/port/`
- Ensure presigned URL constraints (content-type, content-length, TTL) are preserved

## Error Handling

- Return structured errors with appropriate HTTP status codes in API Gateway responses
- Use `slog` for structured logging — always include `fileId`, `userId`, and `status` fields
- Wrap errors with context using `fmt.Errorf("operation: %w", err)`

## Test Patterns

- Unit tests use mock implementations of interfaces defined in `internal/port/`
- No real AWS service calls in tests
- Test files live alongside the code they test (`_test.go` suffix)
- Use table-driven tests for validation logic

## Commit Messages

Follow conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
