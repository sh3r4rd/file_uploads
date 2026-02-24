# DynamoDB table for file upload metadata (Issue #6)
#
# Will include:
# - FileUploads table with fileId as partition key
# - GSI: userId-createdAt-index
# - TTL enabled on ttl attribute
# - On-demand billing mode
