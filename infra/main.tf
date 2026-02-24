# Main Terraform configuration for the Serverless File Upload Service.
# Resource definitions are split across dedicated files (s3.tf, dynamodb.tf, etc.).

locals {
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}
