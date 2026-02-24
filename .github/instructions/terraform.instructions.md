---
applyTo: "infra/**/*.tf"
---

# Terraform Instructions

## File Organization

- Split resources by type: one `.tf` file per AWS service (`s3.tf`, `dynamodb.tf`, `sqs.tf`, `iam.tf`, `lambda.tf`, `apigateway.tf`)
- Variables defined in `variables.tf`
- Outputs defined in `outputs.tf`
- Provider and backend configuration in `main.tf`

## Conventions

- Apply `local.common_tags` to all taggable resources
- Use variables for configurable values — never hardcode account IDs, region, or environment names
- AWS provider version constraint: `~> 5.0`
- Terraform version: `>= 1.5`

## Lambda Configuration

- Runtime: `provided.al2023`
- Architecture: `arm64`
- Each Lambda gets its own IAM role with least-privilege permissions
- Scope IAM policies to specific resource ARNs — never use `*` for resource

## Security

- Never commit `terraform.tfvars` with real values (use `terraform.tfvars.example` as template)
- IAM policies follow least-privilege: specific actions on specific resource ARNs
- S3 bucket policies restrict access to only the required Lambda roles
