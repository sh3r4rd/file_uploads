# IAM roles and policies for Lambda functions (Issue #8)
#
# Will include:
# - RequestUpload Lambda execution role (s3:PutObject, dynamodb:PutItem)
# - ConfirmUpload Lambda execution role (s3:GetObject, s3:DeleteObject, dynamodb:UpdateItem)
# - CloudWatch Logs permissions for both roles
