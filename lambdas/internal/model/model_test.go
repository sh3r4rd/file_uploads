package model_test

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/sh3r4rd/file_uploads/internal/model"
)

func TestUploadRequestJSON(t *testing.T) {
	tests := []struct {
		name  string
		input model.UploadRequest
	}{
		{
			name: "typical request",
			input: model.UploadRequest{
				FileName:      "report.pdf",
				FileSizeBytes: 524288,
				ContentType:   "application/pdf",
			},
		},
		{
			name:  "zero value",
			input: model.UploadRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var got model.UploadRequest
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if got != tt.input {
				t.Errorf("round-trip mismatch: got %+v, want %+v", got, tt.input)
			}
		})
	}
}

func TestUploadRequestJSONFieldNames(t *testing.T) {
	req := model.UploadRequest{
		FileName:      "test.pdf",
		FileSizeBytes: 100,
		ContentType:   "application/pdf",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	for _, key := range []string{"fileName", "fileSizeBytes", "contentType"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q not found", key)
		}
	}
}

func TestUploadResponseJSON(t *testing.T) {
	tests := []struct {
		name  string
		input model.UploadResponse
	}{
		{
			name: "typical response",
			input: model.UploadResponse{
				FileID:    "abc-123",
				UploadURL: "https://s3.amazonaws.com/bucket/key?presigned",
				ExpiresIn: 300,
			},
		},
		{
			name:  "zero value",
			input: model.UploadResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var got model.UploadResponse
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if got != tt.input {
				t.Errorf("round-trip mismatch: got %+v, want %+v", got, tt.input)
			}
		})
	}
}

func TestErrorResponseJSON(t *testing.T) {
	resp := model.ErrorResponse{
		Error:   "VALIDATION_ERROR",
		Message: "file size exceeds 1 MB limit",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got model.ErrorResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got != resp {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, resp)
	}
}

func TestFileMetadataDynamoDB(t *testing.T) {
	meta := model.FileMetadata{
		FileID:        "file-001",
		UserID:        "user-123",
		FileName:      "report.pdf",
		FileSizeBytes: 524288,
		S3Key:         "uploads/user-123/file-001.pdf",
		Status:        model.StatusPending,
		ContentType:   "application/pdf",
		CreatedAt:     "2026-02-25T12:00:00Z",
		UpdatedAt:     "2026-02-25T12:00:00Z",
		TTL:           1740578400,
	}

	av, err := attributevalue.MarshalMap(meta)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}

	var got model.FileMetadata
	if err := attributevalue.UnmarshalMap(av, &got); err != nil {
		t.Fatalf("UnmarshalMap: %v", err)
	}

	if got != meta {
		t.Errorf("round-trip mismatch:\n  got  %+v\n  want %+v", got, meta)
	}
}

func TestFileMetadataDynamoDBAttributeNames(t *testing.T) {
	meta := model.FileMetadata{
		FileID:        "f1",
		UserID:        "u1",
		FileName:      "test.pdf",
		FileSizeBytes: 100,
		S3Key:         "uploads/u1/f1.pdf",
		Status:        model.StatusUploaded,
		ContentType:   "application/pdf",
		CreatedAt:     "2026-01-01T00:00:00Z",
		UpdatedAt:     "2026-01-01T00:00:00Z",
		TTL:           0,
	}

	av, err := attributevalue.MarshalMap(meta)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}

	expected := []string{
		"fileId", "userId", "fileName", "fileSizeBytes",
		"s3Key", "status", "contentType", "createdAt", "updatedAt", "ttl",
	}
	for _, key := range expected {
		if _, ok := av[key]; !ok {
			t.Errorf("expected DynamoDB attribute %q not found", key)
		}
	}
}

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"StatusPending", model.StatusPending, "PENDING"},
		{"StatusUploaded", model.StatusUploaded, "UPLOADED"},
		{"StatusRejected", model.StatusRejected, "REJECTED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("got %q, want %q", tt.got, tt.expected)
			}
		})
	}
}

func TestConstraintConstants(t *testing.T) {
	if model.ContentTypePDF != "application/pdf" {
		t.Errorf("ContentTypePDF = %q, want %q", model.ContentTypePDF, "application/pdf")
	}

	if model.MaxFileSizeBytes != 1_048_576 {
		t.Errorf("MaxFileSizeBytes = %d, want %d", model.MaxFileSizeBytes, 1_048_576)
	}

	if model.PresignedURLTTLSeconds != 300 {
		t.Errorf("PresignedURLTTLSeconds = %d, want %d", model.PresignedURLTTLSeconds, 300)
	}
}
