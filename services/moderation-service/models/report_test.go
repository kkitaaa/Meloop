package models

import (
	"strings"
	"testing"
)

func TestCreateReportRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request CreateReportRequest
		wantErr bool
	}{
		{
			name:    "post with predefined reason",
			request: CreateReportRequest{TargetType: TargetPost, TargetID: "post-1", Reason: "spam"},
		},
		{
			name:    "comment with details for other",
			request: CreateReportRequest{TargetType: TargetComment, TargetID: "comment-1", Reason: "otro", Details: "Contexto"},
		},
		{
			name:    "user target",
			request: CreateReportRequest{TargetType: TargetUser, TargetID: "user-2", Reason: "acoso"},
		},
		{
			name:    "unknown target type",
			request: CreateReportRequest{TargetType: "song", TargetID: "1", Reason: "spam"},
			wantErr: true,
		},
		{
			name:    "unknown reason",
			request: CreateReportRequest{TargetType: TargetPost, TargetID: "post-1", Reason: "inappropriate"},
			wantErr: true,
		},
		{
			name:    "other requires details",
			request: CreateReportRequest{TargetType: TargetPost, TargetID: "post-1", Reason: "otro"},
			wantErr: true,
		},
		{
			name:    "details max length counts unicode characters",
			request: CreateReportRequest{TargetType: TargetPost, TargetID: "post-1", Reason: "spam", Details: strings.Repeat("á", 501)},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
