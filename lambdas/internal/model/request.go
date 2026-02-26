package model

// UploadRequest is the JSON body sent by clients to POST /uploads.
type UploadRequest struct {
	FileName      string `json:"fileName"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	ContentType   string `json:"contentType"`
}
