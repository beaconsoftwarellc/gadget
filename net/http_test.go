package net

import (
	"context"
	"net/http"
	"testing"
	"time"

	assert1 "github.com/stretchr/testify/assert"
)

func Test_client_DoWithContext(t *testing.T) {
	assert := assert1.New(t)
	client := NewHTTPRedirectClient(time.Minute)
	request, err := http.NewRequest("GET", "http://localhost", nil)
	assert.NoError(err)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Nanosecond)
	cancel()
	_, err = client.DoWithContext(ctx, request)
	assert.EqualError(err, "request to http://localhost was cancelled by controlling context")
}
