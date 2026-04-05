[![Certified Shovelware](https://justin.searls.co/img/shovelware.svg)](https://justin.searls.co/shovelware/)

# sonaveeb-cli

CLI for querying Estonian word forms from the [Ekilex API](https://github.com/keeleinstituut/ekilex/wiki/Ekilex-API).

## Installation

```sh
go install github.com/LarsEckart/sonaveeb-cli@latest
```

Or build from source:

```sh
make build
```

Available helper targets:

```sh
make fmt
make test
make check
```

## Releasing

1. Make sure `main` is clean and up to date.
2. Run the release checks:

```sh
go build && go test
golangci-lint run
```

3. Create and push a version tag:

```sh
git checkout main
git pull
git tag v0.1.7
git push origin main
git push origin v0.1.7
```

Pushing a `v*` tag triggers GitHub Actions to build and publish a release via GoReleaser.

To install that exact release locally:

```sh
go install github.com/LarsEckart/sonaveeb-cli@v0.1.7
hash -r
sonaveeb-cli --version
```

## Configuration

Get an API key from your [Ekilex profile page](https://ekilex.ee).

Precedence (highest to lowest):
1. Environment variable
2. Local config file (`./config`)
3. User config file (`~/.config/sonaveeb/config`)

### Environment variable
```sh
export EKILEX_API_KEY="your-key-here"
```

### Config file
Just the API key on a single line (no formatting):
```sh
echo "your-key-here" > config                       # local
echo "your-key-here" > ~/.config/sonaveeb/config    # user
```

## Usage

```sh
sonaveeb-cli [flags] <word>
```

### Flags

- `-json` - Output raw JSON from API
- `-all` - Show all forms (not just key forms)
- `-homonym=N` - Select which homonym to show (when multiple exist)
- `-q`, `-quiet` - Minimal output (forms only)
- `--version` - Print version
- `-h` - Show help

### Examples

```sh
# Noun - shows key forms
sonaveeb-cli puu
# puu (noun, type 26)
#   ainsuse nimetav:                    puu
#   ainsuse omastav:                    puu
#   ainsuse osastav:                    puud
#   mitmuse osastav:                    puusid

# Verb - shows key forms
sonaveeb-cli tegema
# tegema (verb, type 28)
#   ma-tegevusnimi:                     tegema
#   da-tegevusnimi:                     teha
#   kindel kõneviis olevikus 3.p:       teeb
#   mineviku kesksõna umbisikuline:     tehtud

# Word with multiple homonyms
sonaveeb-cli pank
# pank (noun, type 22)  [1 of 3 — use --homonym=N for others]
#   ainsuse nimetav:                    pank
#   ...

# Select specific homonym
sonaveeb-cli --homonym=2 pank

# All forms
sonaveeb-cli -all puu

# JSON output
sonaveeb-cli -json puu

# Version
sonaveeb-cli --version
```
