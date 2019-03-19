# Pommel
## Use Vault as a Blob Storage Service, WIP!

Pommel is an S3ish interface for Vault values.

Vault is great for storing secrets, but it's kind of annoying to manually assert types on each value of [the Vault API's `map[string]interface{}`](https://godoc.org/github.com/hashicorp/vault/api#Secret). Pommel asks for a `bucket` (a Vault path) and a `key` (a Vault key) and returns the Vault value as a `[]byte`. From there, you can decode the blob however you like.

## Examples
Pommel does not provide tokens, so users must log in or provide their own authentication tokens! It will, however, use `$VAULT_ADDR` and `~/.vault-token` as defaults if no values are explicitly passed.

### CLI
1. `make bin` or `go build ./cmd/pommel`.
2. `./pommel read -a="myvault.com" -t="~/.vault-token" -b="path/to/secret" -k="key"` or, if you want to use defaults, `./pommel -b="path/to/secret" -k="key"`.
3. You'll then be prompted to print your secret to stdout!

### API
```go
func main() {
    // Use user defaults.
    pom := pommel.NewClient(nil)
    err := pom.Get(context.Background(), "fake", "even_faker.json")
	if err != nil {
		panic(err)
	}
	raw := bytes.NewBuffer(bb)
	var cfg resolver.Config // Some application config
	if err := json.NewDecoder(raw).Decode(&cfg); err != nil {
		panic(errors.Wrap(err, "could not decode secret"))
    }
    fmt.Println(cfg)
}
```
## Roadmap:
Pommel is very much a WIP and was created to satisfy a specific workflow: retrieving JSON encoded values from Vault. The goal is to expand the functionality to match an S3 like interface in a CLI and API. It currently supports:
* Reading :blue_book:
  
Next up:

* Writing :pen:

### Related Alternatives and Prior Art
[MapStructure](https://github.com/mitchellh/mapstructure) marshals a map[string]interface{} into a struct, which, bo be honest, is probably sufficient for most use cases.
[Viper](https://github.com/spf13/viper) essentially does the same thing...but only for etcd and Consul. I didn't realize this until I started using Cobra for the CLI...