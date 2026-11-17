package errors

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTagStatus(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		base        *status.Status
		wantCode    codes.Code
		wantMessage string
	}{
		{
			name:        "not found",
			tag:         "store",
			base:        status.New(codes.NotFound, "missing"),
			wantCode:    codes.NotFound,
			wantMessage: "[store]: rpc error: code = NotFound desc = missing",
		},
		{
			name:        "ok",
			tag:         "store",
			base:        status.New(codes.OK, "ignored"),
			wantCode:    codes.OK,
			wantMessage: "",
		},
		{
			name:        "empty tag",
			tag:         "",
			base:        status.New(codes.PermissionDenied, "denied"),
			wantCode:    codes.PermissionDenied,
			wantMessage: "[]: rpc error: code = PermissionDenied desc = denied",
		},
		{
			name:        "empty message",
			tag:         "service",
			base:        status.New(codes.Internal, ""),
			wantCode:    codes.Internal,
			wantMessage: "[service]: rpc error: code = Internal desc = ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TagStatus(tt.tag, tt.base)
			if tt.wantCode == codes.OK {
				if got != nil {
					t.Fatalf("Code() = %s, want <nil>", got.GRPCStatus().Code())
				} else {
					return
				}
			}
			gotStatus := got.GRPCStatus()
			if gotStatus.Code() != tt.wantCode {
				t.Fatalf("Code() = %v, want %v", gotStatus.Code(), tt.wantCode)
			}
			if gotStatus.Message() != tt.wantMessage {
				t.Fatalf("Message() = %q, want %q", gotStatus.Message(), tt.wantMessage)
			}
		})
	}
}
