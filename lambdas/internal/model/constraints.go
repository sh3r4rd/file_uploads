package model

// Domain constants shared across handler, validation, and storage packages.
const (
	ContentTypePDF         = "application/pdf"
	MaxFileSizeBytes       = int64(1_048_576) // 1 MB
	PresignedURLTTLSeconds = 300              // 5 minutes
)
