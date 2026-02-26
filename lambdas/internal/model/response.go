package model

// UploadResponse is returned on a successful POST /uploads request.
type UploadResponse struct {
	FileID    string `json:"fileId"`
	UploadURL string `json:"uploadUrl"`
	ExpiresIn int    `json:"expiresIn"`
}

// ErrorResponse is returned for any failed API request.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
