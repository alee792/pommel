package cli

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/pflag"

	"github.com/pkg/errors"

	"github.com/alee792/pommel"
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
	Put(ctx context.Context, r io.Reader, bucket, key string) error
}

// Flags from the CLI.
type Flags struct {
	Addr       string `arg:"-a" help:"vault addr"`
	TokenPath  string `arg:"-p" help:"path to token"`
	Token      string `arg:"-t" help:"vault token"`
	Bucket     string `arg:"-b,required" help:"path to value"`
	Key        string `arg:"-k,required" help:"key for value"`
	HidePrompt bool
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

// Get returns reader from CLI flags.
func Get() (io.Reader, error) {
	hilt := NewHilt()
	// Required flags. Will be replaced by args.
	pflag.StringVarP(&hilt.Bucket, "bucket", "b", "", "A path in Vault.")
	pflag.StringVarP(&hilt.Key, "key", "k", "", "A key in Vault.")

	// Optional flags.
	pflag.StringVarP(&hilt.Addr, "addr", "a", "", "Address of Vault server.")
	pflag.StringVarP(&hilt.Token, "tkn", "t", "", "Vault auth token.")
	pflag.StringVarP(&hilt.TokenPath, "tknp", "p", "~/.vault-token", "Path to Vault auth token.")
	pflag.BoolVarP(&hilt.HidePrompt, "hide", "h", false, "Hide prompt to print to stdout.")

	pflag.Parse()
	// Additional defaults.

	if hilt.Bucket == "" || hilt.Key == "" {
		return nil, errors.New("must provide bucket and key")
	}

	if err := hilt.AddVault(); err != nil {
		return nil, errors.Wrap(err, "vault setup failed")
	}

	raw, err := hilt.Provider("vault").Client.Get(context.Background(), hilt.Bucket, hilt.Key)
	if err != nil {
		return nil, errors.Wrap(err, "Get failed")
	}
	return raw, nil
}

// AddVault as a provider.
func (h *Hilt) AddVault() error {
	if h.Addr == "" {
		h.Addr = os.Getenv("VAULT_ADDR")
	}

	cfg, err := CreateConfig(h.Flags)
	if err != nil {
		return errors.Wrap(err, "Config creation failed")
	}
	client, err := pommel.NewClient(cfg)
	if err != nil {
		return errors.Wrap(err, "Vault client creation failed")
	}
	h.AddProvider(&Provider{
		Scheme: "vault",
		Client: client,
	})
	return nil
}

// Schemes returns the Hilt's current schemes.
// See comments on Provider.
func (h *Hilt) Schemes() []string {
	return h.schemes
}

// Provider returns a Provider for a given scheme.
// This abstraction prevents consumers from fiddling
// with thread-unfsafe internals.
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

// CreateConfig from Args and attempt to set default variables
// from a user's enivronment.
func CreateConfig(f *Flags) (*pommel.Config, error) {
	if f.TokenPath == "" {
		f.TokenPath = "~/.vault-token"
	}

	if f.Token == "" {
		tkn, err := GetToken(f.TokenPath)
		if err != nil {
			return nil, err
		}
		f.Token = tkn
	}

	if f.Addr == "" {
		f.Addr = os.Getenv("VAULT_ADDR")
	}
	cfg := &pommel.Config{
		Addr:  f.Addr,
		Token: f.Token,
	}
	return cfg, nil
}

// ValidateSrcDst to ensure at least one valid remote URI.
// We're not in the business of local file managment here!
func ValidateSrcDst(h *Hilt, args []string) error {
	if len(args) != 2 {
		return errors.New("requires exactly two args")
	}
	// Verbose logic for verbose errors.
	if !hasValidPrefix(args[0], h.Schemes()) && !hasValidPrefix(args[1], h.Schemes()) {
		return errors.New("requires valid URI")
	}
	return nil
}

// GetToken from local file system.
func GetToken(tokenPath string) (string, error) {
	// Expand "~" to absolute path.
	if strings.Contains(tokenPath, "~") {
		usr, _ := user.Current()
		tokenPath = strings.Replace(tokenPath, "~", usr.HomeDir, -1)
	}
	tkn, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return "", errors.Wrapf(err, "invalid token path %s", tokenPath)
	}
	return string(tkn), nil
}

func hasValidPrefix(s string, pp []string) bool {
	for _, p := range pp {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// ParseURI for its components from a string.
func ParseURI(uri string) (schemes, bucket, key string, err error) {
	// Try local.
	scheme, bucket, key, ok := ParseLocal(uri)
	if ok {
		return scheme, bucket, key, nil
	}
	// Try remote.
	sep := "://"
	ss := strings.Split(uri, sep)
	if len(ss) != 2 {
		return "", "", "", errors.New("invalid uri")
	}

	scheme, path := ss[0], ss[1]
	bucket, key = filepath.Split(path)
	if bucket == "" || key == "" {
		return "", "", "", errors.New("bucket and key required")
	}
	return scheme, bucket, key, nil
}

// ParseLocal fields from a possible local URI.
func ParseLocal(uri string) (s, b, k string, ok bool) {
	if _, err := os.Stat(uri); err != nil {
		return "", "", "", false
	}
	b, k = filepath.Split(uri)
	return "local", b, k, true
}
