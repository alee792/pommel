// Package local is a local filesystem client.
package local

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
)

// Client for local filesystem interaction.
type Client struct {
}

// Get reader from local file.
func (c *Client) Get(ctx context.Context, bucket, key string) (io.Reader, error) {
	ioutil.ReadFile(fmt.Sprintf("%s/%s", bucket, key))
	return nil, nil
}

// Put reader to local file.
func (c *Client) Put(ctx context.Context, r io.Reader, bucket, key string) error {
	return nil
}

// Writer for local files.
type Writer struct {
	path string
	*Client
}

// Write to bucket location.
func (w *Writer) Write(p []byte) (int, error) {
	err := ioutil.WriteFile(w.path, p, 0644)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Writer is created at a bucket and key.
func (c *Client) Writer(ctx context.Context, bucket, key string) *Writer {
	return &Writer{
		path:   fmt.Sprintf("%s/%s", bucket, key),
		Client: c,
	}
}
