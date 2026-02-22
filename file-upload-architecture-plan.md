# Serverless File Upload System — Architecture & Implementation Plan

## 1. Architecture Overview

The system uses four AWS services working together in a request flow:

```
Client (HTTPS)
   │
   ▼
API Gateway (REST API)
   │  ── API Key authentication
   │  ── Request validation
   ▼
Lambda (Go)
   │  ── Validates file (type, size)
   │  ── Generates S3 presigned URL
   │  ── Writes metadata to DynamoDB
   ▼
S3 (storage)          DynamoDB (metadata)
   └─ /uploads/{userId}/{uuid}.pdf    └─ FileUploads table
```

There is an important architectural decision to call out upfront: **direct upload vs. presigned URL**. The two approaches are:

| Approach | How it works | Pros | Cons |
|---|---|---|---|
| **Direct (passthrough)** | Client → API GW → Lambda → S3 | Simpler client logic | Lambda pays for all data transfer; API GW has a 10 MB payload limit; Lambda timeout risk on slow uploads |
| **Presigned URL (two-step)** | Client → Lambda (get URL) → S3 (direct PUT) | Lambda only handles lightweight JSON; no payload limits beyond your own; S3 handles the heavy lifting | Slightly more complex client; requires a second HTTP call |

**Recommendation: Presigned URL.** Even with a 1 MB limit today, the presigned URL approach is architecturally superior. Lambda invocations stay fast and cheap (~50ms), you avoid base64-encoding overhead through API Gateway, and if the size limit ever increases, you don't need to re-architect. The remainder of this plan uses this approach.

---

## 2. Detailed Request Flow

### Step 1 — Request Upload (Client → API Gateway → Lambda)

```
POST /uploads
Headers:
  x-api-key: <api-key>
  Content-Type: application/json
Body:
  {
    "fileName": "quarterly-report.pdf",
    "fileSizeBytes": 524288,
    "contentType": "application/pdf"
  }
```

The Lambda function:

1. Validates the request payload (file name present, content type is `application/pdf`, size ≤ 1,048,576 bytes).
2. Generates a UUID for the file (`fileId`).
3. Constructs the S3 object key: `uploads/{userId}/{fileId}.pdf`.
4. Generates a presigned PUT URL for that key with a short TTL (5 minutes) and conditions:
   - `Content-Type` must be `application/pdf`
   - `Content-Length` must be ≤ 1,048,576 bytes
5. Writes a metadata record to DynamoDB with status `PENDING`.
6. Returns the presigned URL and `fileId` to the client.

**Response:**

```json
{
  "fileId": "a1b2c3d4-...",
  "uploadUrl": "https://bucket.s3.amazonaws.com/uploads/...?X-Amz-...",
  "expiresIn": 300
}
```

### Step 2 — Upload File (Client → S3)

The client PUTs the raw PDF bytes directly to S3 using the presigned URL. S3 enforces the presigned conditions (content type, size). No Lambda or API Gateway involvement here.

```
PUT <presigned-url>
Headers:
  Content-Type: application/pdf
Body: <raw PDF bytes>
```

### Step 3 — Confirm Upload (S3 Event → Lambda)

An S3 event notification (`s3:ObjectCreated:Put`) triggers a second Lambda function that:

1. Verifies the object exists and its size is ≤ 1 MB (defense in depth).
2. Optionally validates the file is a real PDF (check magic bytes `%PDF-`).
3. Updates the DynamoDB record status from `PENDING` → `UPLOADED`.
4. If validation fails, deletes the object from S3 and sets status to `REJECTED`.

---

## 3. AWS Resource Design

### 3.1 S3 Bucket

**Bucket name:** `{project-name}-file-uploads-{account-id}-{region}`

**Configuration:**
- Versioning: Enabled (protects against accidental overwrites/deletes).
- Encryption: SSE-S3 (AES-256) at rest — sufficient for most use cases; upgrade to SSE-KMS if you need audit trails on key usage.
- Public access: Blocked entirely (all four block settings enabled).
- CORS: Configured to allow PUT from your web application's origin (needed for presigned URL uploads from the browser).
- Lifecycle rule: Transition `PENDING` objects older than 24 hours to deletion (clean up abandoned uploads).

**Key structure:**
```
uploads/{userId}/{fileId}.pdf
```

Partitioning by `userId` keeps things organized and makes it straightforward to implement per-user access controls or listing later.

### 3.2 DynamoDB Table

**Table name:** `FileUploads`

| Attribute | Type | Role |
|---|---|---|
| `fileId` (PK) | String (UUID) | Partition key |
| `userId` (GSI PK) | String | Who uploaded it |
| `fileName` | String | Original file name from client |
| `fileSizeBytes` | Number | Declared file size |
| `s3Key` | String | Full S3 object key |
| `status` | String | `PENDING` / `UPLOADED` / `REJECTED` |
| `contentType` | String | MIME type (`application/pdf`) |
| `createdAt` | String (ISO 8601) | Upload request timestamp |
| `updatedAt` | String (ISO 8601) | Last status change |
| `ttl` | Number (epoch) | DynamoDB TTL — auto-delete `PENDING` records after 24h |

**GSI:** `userId-createdAt-index` (PK: `userId`, SK: `createdAt`) — enables querying "all uploads by user X, most recent first."

**Capacity:** On-demand (pay-per-request). At low-to-moderate volume this is cheaper and eliminates capacity planning. Switch to provisioned with auto-scaling if you hit predictable, sustained throughput.

### 3.3 Lambda Functions

Two functions, both written in Go and compiled to `provided.al2023` runtime (Amazon Linux 2023, ARM64 for cost savings):

#### `RequestUploadHandler`

- **Trigger:** API Gateway POST /uploads
- **Memory:** 128 MB (this is lightweight JSON processing)
- **Timeout:** 10 seconds
- **Environment variables:** `BUCKET_NAME`, `TABLE_NAME`, `UPLOAD_URL_TTL`
- **IAM permissions:**
  - `s3:PutObject` on `arn:aws:s3:::{bucket}/uploads/*`
  - `dynamodb:PutItem` on the FileUploads table

#### `ConfirmUploadHandler`

- **Trigger:** S3 `s3:ObjectCreated:Put` on prefix `uploads/`
- **Memory:** 256 MB (reads object head / first bytes)
- **Timeout:** 30 seconds
- **Environment variables:** `TABLE_NAME`
- **IAM permissions:**
  - `s3:GetObject`, `s3:DeleteObject` on `arn:aws:s3:::{bucket}/uploads/*`
  - `dynamodb:UpdateItem` on the FileUploads table

### 3.4 API Gateway

**Type:** REST API (not HTTP API — REST API has native API key support with usage plans).

**Resources:**
```
POST /uploads  →  RequestUploadHandler
```

**Configuration:**
- API key required: `true` on the POST method.
- Usage plan: Rate limit of 10 req/sec, burst 20, quota 1000/day (adjust to your needs).
- Request validation: Enable body validation with a JSON schema model that rejects payloads without `fileName` and `contentType`.
- Stage: `prod` with access logging enabled to CloudWatch.

**Authentication note:** The PRD specifies "API keys for authentication." API keys are fine for rate limiting and identifying callers, but they are **not** a strong authentication mechanism (they're passed in headers, easily leaked). For production, strongly consider adding a Lambda authorizer or Cognito user pool authorizer in front of this. The `userId` should come from a verified token, not from the client payload. For now, we proceed with API keys as specified, but this is a known gap to address before a production launch.

---

## 4. Security Considerations

### Authentication & Authorization
- API Gateway API keys gate access to the upload endpoint.
- Presigned URLs are short-lived (5 min) and scoped to a specific S3 key and content type.
- **Future improvement:** Replace or augment API keys with JWT-based auth (Cognito or custom authorizer) so `userId` is derived from a verified token rather than trusted from the client.

### Data in Transit
- API Gateway enforces HTTPS by default (no HTTP endpoint exposed).
- S3 presigned URLs use HTTPS.

### Data at Rest
- S3 server-side encryption (SSE-S3).
- DynamoDB encryption at rest (enabled by default).

### Input Validation (Defense in Depth)
Validation happens at **three layers:**

1. **Client-side** (future web UI): Check file extension and size before upload.
2. **Lambda (RequestUploadHandler):** Validate declared content type and size.
3. **Lambda (ConfirmUploadHandler):** Validate actual object size and PDF magic bytes after upload to S3.

### Least Privilege IAM
Each Lambda gets its own role with only the permissions it needs, scoped to specific resource ARNs.

---

## 5. Implementation Plan

### Phase 1 — Infrastructure (IaC with CDK or Terraform)

Define all resources in code. I'd recommend **AWS CDK in TypeScript** for this since you're comfortable with TS and it gives you strong typing and good abstractions for serverless stacks. Alternatively, Terraform with HCL works well if that's your team's standard.

**Resources to define:**
1. S3 bucket with encryption, versioning, public access block, CORS, lifecycle rules.
2. DynamoDB table with GSI and TTL enabled.
3. IAM roles for each Lambda (scoped permissions).
4. Lambda functions (Go binaries packaged as zip, ARM64).
5. API Gateway REST API with API key, usage plan, request validator, and Lambda integration.
6. S3 event notification to trigger the confirmation Lambda.
7. CloudWatch log groups with retention policies.

### Phase 2 — Lambda: RequestUploadHandler (Go)

```
cmd/
  request-upload/
    main.go          # Lambda entrypoint, handler wiring
internal/
  handler/
    request_upload.go  # Core handler logic
  validation/
    validate.go        # File validation rules
  storage/
    s3.go              # Presigned URL generation
    dynamodb.go        # Metadata writes
  model/
    types.go           # Request/response structs, DynamoDB item
```

**Key implementation details:**

- Use `github.com/aws/aws-lambda-go` for the Lambda handler.
- Use `github.com/aws/aws-sdk-go-v2` (v2 SDK — not the legacy v1).
- Presigned URL generation via `s3.NewPresignClient`.
- Set presigned URL conditions: content type and content-length-range.
- Generate UUIDs with `github.com/google/uuid`.
- Return structured API Gateway proxy responses with appropriate status codes (200, 400, 413, 500).

**Validation rules:**
- `contentType` must be exactly `application/pdf`.
- `fileName` must end with `.pdf` (case-insensitive).
- `fileSizeBytes` must be > 0 and ≤ 1,048,576.
- `fileName` must not contain path traversal characters (`..`, `/`, `\`).

### Phase 3 — Lambda: ConfirmUploadHandler (Go)

```
cmd/
  confirm-upload/
    main.go
internal/
  handler/
    confirm_upload.go
  validation/
    pdf.go              # Magic byte validation
  storage/
    s3.go
    dynamodb.go
```

**Key implementation details:**

- Parse S3 event notification from the Lambda event payload.
- `HeadObject` to get actual content length.
- `GetObject` with range header (`Range: bytes=0-4`) to read the first 5 bytes and verify the PDF magic number (`%PDF-`).
- Update DynamoDB status to `UPLOADED` or `REJECTED`.
- If rejected, delete the object from S3.

### Phase 4 — Testing

**Unit tests (Go):**
- Validation logic (accepted/rejected file types, sizes, names).
- Handler logic with mocked AWS SDK clients (use interfaces + mock implementations).

**Integration tests:**
- Deploy to a staging environment.
- Use the AWS SDK (or `curl`) to call the API Gateway endpoint, receive a presigned URL, upload a valid PDF, and verify the DynamoDB record transitions to `UPLOADED`.
- Upload invalid files (wrong type, oversized) and verify rejection.
- Upload with an expired/invalid API key and verify 403.

**Load test (optional):**
- Use a tool like `k6` or `artillery` to verify the system handles concurrent uploads gracefully.

### Phase 5 — Observability

- **Structured logging:** Use `slog` (Go 1.21+) in Lambda functions. Log `fileId`, `userId`, and `status` on every operation for traceability.
- **CloudWatch Metrics:** Monitor Lambda duration, error count, throttles. Monitor API Gateway 4xx/5xx rates.
- **Alarms:** Alert on elevated 5xx error rates or Lambda error percentages.
- **X-Ray (optional):** Enable tracing on API Gateway and Lambda for end-to-end request visibility.

---

## 6. Cost Estimate (Low Volume — ~1,000 uploads/month)

| Service | Usage | Estimated Cost |
|---|---|---|
| API Gateway | 1,000 requests | ~$0.0035 |
| Lambda (RequestUpload) | 1,000 invocations × 128 MB × 100ms | ~$0.0002 |
| Lambda (ConfirmUpload) | 1,000 invocations × 256 MB × 200ms | ~$0.0008 |
| S3 | 1 GB stored, 1,000 PUTs, 1,000 GETs | ~$0.03 |
| DynamoDB | 1,000 writes, 1,000 reads (on-demand) | ~$0.003 |
| **Total** | | **< $1/month** (within free tier for first 12 months) |

The serverless model means you pay essentially nothing at low volume. Costs scale linearly and stay very manageable even at 100k uploads/month (~$5–10).

---

## 7. Project Structure

```
file-upload-service/
├── infra/                    # CDK or Terraform
│   ├── lib/
│   │   └── upload-stack.ts   # All resource definitions
│   ├── bin/
│   │   └── app.ts
│   ├── cdk.json
│   └── package.json
├── lambdas/
│   ├── cmd/
│   │   ├── request-upload/
│   │   │   └── main.go
│   │   └── confirm-upload/
│   │       └── main.go
│   ├── internal/
│   │   ├── handler/
│   │   │   ├── request_upload.go
│   │   │   └── confirm_upload.go
│   │   ├── validation/
│   │   │   ├── validate.go
│   │   │   ├── validate_test.go
│   │   │   ├── pdf.go
│   │   │   └── pdf_test.go
│   │   ├── storage/
│   │   │   ├── s3.go
│   │   │   └── dynamodb.go
│   │   └── model/
│   │       └── types.go
│   ├── go.mod
│   └── go.sum
├── scripts/
│   ├── build.sh              # Cross-compile Go for AL2023 ARM64
│   └── deploy.sh             # Build + cdk deploy
├── Makefile
└── README.md
```

---

## 8. Sequence Diagram

```
 Client              API Gateway          RequestUpload λ        S3                ConfirmUpload λ      DynamoDB
   │                     │                      │                 │                      │                 │
   │ POST /uploads       │                      │                 │                      │                 │
   │ (fileName, size)    │                      │                 │                      │                 │
   │────────────────────>│                      │                 │                      │                 │
   │                     │ Validate API key     │                 │                      │                 │
   │                     │ Validate body schema │                 │                      │                 │
   │                     │─────────────────────>│                 │                      │                 │
   │                     │                      │ Validate input  │                      │                 │
   │                     │                      │ Generate fileId │                      │                 │
   │                     │                      │                 │                      │                 │
   │                     │                      │ PutItem(PENDING)│                      │                 │
   │                     │                      │────────────────────────────────────────────────────────>│
   │                     │                      │                 │                      │                 │
   │                     │                      │ Presign PutObj  │                      │                 │
   │                     │                      │────────────────>│                      │                 │
   │                     │                      │<────────────────│                      │                 │
   │                     │                      │                 │                      │                 │
   │                     │<─────────────────────│                 │                      │                 │
   │<────────────────────│  {uploadUrl, fileId} │                 │                      │                 │
   │                     │                      │                 │                      │                 │
   │ PUT <presigned-url> │                      │                 │                      │                 │
   │ (raw PDF bytes)     │                      │                 │                      │                 │
   │────────────────────────────────────────────────────────────>│                      │                 │
   │                     │                      │                 │ 200 OK               │                 │
   │<────────────────────────────────────────────────────────────│                      │                 │
   │                     │                      │                 │                      │                 │
   │                     │                      │                 │ S3 Event Notification│                 │
   │                     │                      │                 │─────────────────────>│                 │
   │                     │                      │                 │                      │ HeadObject      │
   │                     │                      │                 │                      │ GetObject(0-4)  │
   │                     │                      │                 │<─────────────────────│                 │
   │                     │                      │                 │─────────────────────>│                 │
   │                     │                      │                 │                      │ Validate        │
   │                     │                      │                 │                      │ UpdateItem      │
   │                     │                      │                 │                      │ (→ UPLOADED)    │
   │                     │                      │                 │                      │────────────────>│
   │                     │                      │                 │                      │                 │
```

---

## 9. Open Questions / Future Considerations

1. **Authentication upgrade:** API keys are a weak authentication mechanism. Plan to introduce Cognito or a custom JWT authorizer so that `userId` is derived from a verified token, not from an untrusted client payload.

2. **Upload progress / status polling:** The client currently has no way to know when the ConfirmUpload Lambda has finished processing. Options: poll a `GET /uploads/{fileId}` endpoint, or use WebSockets via API Gateway for real-time status.

3. **Virus scanning:** For production, consider triggering a ClamAV scan (via a Lambda layer or container image) on uploaded files before marking them `UPLOADED`.

4. **File processing pipeline:** The PRD mentions files will be "later analyzed and processed." When that's ready, the ConfirmUpload Lambda (or a separate post-processing Lambda) can push to an SQS queue or EventBridge for downstream consumers.

5. **Multi-file uploads:** The current design handles one file per request. If batch uploads are needed, the endpoint can return multiple presigned URLs in one response.

6. **Retry / idempotency:** S3 event notifications are at-least-once. The ConfirmUpload Lambda should be idempotent — updating a record to `UPLOADED` twice is harmless, but be careful with any side effects.

7. **Web interface:** Deferred per the PRD. When built, it will need the S3 CORS configuration and should use a multipart form or the Fetch API to PUT directly to the presigned URL.
