package cli

import (
	"context"
	"io"
	"sync"
)

// Hilt allows Provider and CLI dependencies to be shared and reused.
type Hilt struct {
	*Flags
	// e.g. providers["vault"] = &pommel.Client{}
	providers map[string]*Provider
	schemes   []string
	mux       *sync.Mutex
}

// Provider of remote configurations and secrets.
type Provider struct {
	Scheme string
	Client Client
}

// Client defines a remote store's expected
// capabilities with an S3-like interface.
type Client interface {
	Get(ctx context.Context, bucket, key string) (io.Reader, error)
	// Put(ctx context.Context, bucket, key string, body io.Reader) error
}

// Flags from the CLI.
type Flags struct {
	Addr      string `arg:"-a" help:"vault addr"`
	TokenPath string `arg:"-p" help:"path to token"`
	Token     string `arg:"-t" help:"vault token"`
	Bucket    string `arg:"-b,required" help:"path to value"`
	Key       string `arg:"-k,required" help:"key for value"`
}

// NewHilt creates a Hilt with providers, shared flags and validators.
// Pommeler and Flags are empty because the RootCmd resolves them.
func NewHilt() *Hilt {
	h := &Hilt{
		Flags:     &Flags{},
		providers: make(map[string]*Provider),
		mux:       &sync.Mutex{},
	}
	return h
}

// Provider returns a Provider for a given scheme.
// This abstraction prevents consumers from fiddling
// with thread-unfsafe internals, like `providers`.
func (h *Hilt) Provider(scheme string) *Provider {
	return h.providers[scheme]
}

// AddProvider to Hilt's Provider map and schemes.
func (h *Hilt) AddProvider(p *Provider) {
	h.mux.Lock()
	defer h.mux.Unlock()
	h.providers[p.Scheme] = p
	h.schemes = append(h.schemes, p.Scheme)
}

// RemoveProvider from Hilt.
func (h *Hilt) RemoveProvider(scheme string) {
	h.mux.Lock()
	defer h.mux.Unlock()
	for i, s := range h.schemes {
		if s == scheme {
			h.schemes = append(h.schemes[:i], h.schemes[:i+1]...)
			return
		}
	}
	delete(h.providers, scheme)
}
