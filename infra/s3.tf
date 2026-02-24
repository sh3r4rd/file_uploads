# S3 bucket for file uploads (Issue #5)
#
# Will include:
# - Bucket with versioning enabled
# - SSE-S3 encryption
# - Public access block (all four settings)
# - CORS configuration for browser uploads
# - Lifecycle rule to clean up PENDING objects after 24h
# - Event notification to trigger ConfirmUpload Lambda
