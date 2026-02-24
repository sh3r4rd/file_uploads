# Lambda function definitions (Issue #9)
#
# Will include:
# - RequestUploadHandler (128 MB, 10s timeout, API GW trigger)
# - ConfirmUploadHandler (256 MB, 30s timeout, S3 event trigger)
# - Both using provided.al2023 runtime with ARM64 architecture
# - Environment variables: BUCKET_NAME, TABLE_NAME, UPLOAD_URL_TTL
