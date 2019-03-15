package pommel_test

import (
	"os"
	"testing"

	"github.com/alee792/pommel"
	"github.com/google/go-cmp/cmp"
)

func TestNewClient(t *testing.T) {

}

func TestAutoConfig(t *testing.T) {
	var (
		addrEnvVar = "VAULT_ADDR"
		addr       = "test.com"
	)
	curAddr := os.Getenv(addrEnvVar)
	defer os.Setenv(addrEnvVar, curAddr)

	// AutoConfig does not guarantee a token!
	os.Setenv(addrEnvVar, addr)
	want := &pommel.Config{
		Addr: "test.com",
	}
	got, err := pommel.AutoConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Errorf("config does not match: %+v", cmp.Diff(want, got))
	}
}
