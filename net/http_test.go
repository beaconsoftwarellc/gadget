package net

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_client_DoWithContext(t *testing.T) {
	assert := assert.New(t)
	client := NewHTTPRedirectClient(time.Minute)
	request, err := http.NewRequest("GET", "http://localhost", nil)
	assert.NoError(err)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Nanosecond)
	cancel()
	_, err = client.DoWithContext(ctx, request)
	assert.EqualError(err, "request to http://localhost was cancelled by controlling context")
}
