// Package example is a template for new providers.
package example

import (
	"context"
	"io"
)

// Client for provider interaction.
type Client struct {
}

// Get from provider.
func (c *Client) Get(ctx context.Context, bucket, key string) (io.Reader, error) {
	return nil, nil
}

// Put to provider.
func (c *Client) Put(ctx context.Context, r io.Reader, bucket, key string) error {
	return nil
}

// Writer for local files.
type Writer struct {
	bucket string
	key    string
	*Client
}

// Write to bucket location.
func (w *Writer) Write(p []byte) (int, error) {
	return 0, nil
}

// Writer is created at a bucket and key.
func (c *Client) Writer(ctx context.Context, bucket, key string) *Writer {
	return nil
}
