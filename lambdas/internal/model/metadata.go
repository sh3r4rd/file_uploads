package model

// FileMetadata represents a single item in the FileUploads DynamoDB table.
type FileMetadata struct {
	FileID        string `dynamodbav:"fileId"`
	UserID        string `dynamodbav:"userId"`
	FileName      string `dynamodbav:"fileName"`
	FileSizeBytes int64  `dynamodbav:"fileSizeBytes"`
	S3Key         string `dynamodbav:"s3Key"`
	Status        string `dynamodbav:"status"`
	ContentType   string `dynamodbav:"contentType"`
	CreatedAt     string `dynamodbav:"createdAt"`
	UpdatedAt     string `dynamodbav:"updatedAt"`
	TTL           int64  `dynamodbav:"ttl"`
}

// Status constants for FileMetadata.Status.
const (
	StatusPending  = "PENDING"
	StatusUploaded = "UPLOADED"
	StatusRejected = "REJECTED"
)
