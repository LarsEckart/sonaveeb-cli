# sonaveeb-cli

CLI for querying Estonian word forms from the [Ekilex API](https://github.com/keeleinstituut/ekilex/wiki/Ekilex-API).

## Installation

```sh
go install github.com/lars/sonaveeb-cli@latest
```

Or build from source:

```sh
go build -o sonaveeb .
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
sonaveeb <word> [flags]
```

### Flags

- `-json` - Output raw JSON from API
- `-all` - Show all forms (not just key forms)
- `-q`, `-quiet` - Minimal output (forms only)
- `-version` - Print version
- `-h` - Show help

### Examples

```sh
# Noun - shows SgN, SgG, SgP, PlP
sonaveeb puu
# puu (noun, type 26)
#   SgN:   puu
#   SgG:   puu
#   SgP:   puud
#   PlP:   puid

# Verb - shows ma, da, 3sg, tud
sonaveeb tegema
# tegema (verb, type 28)
#   ma:    tegema
#   da:    teha
#   3sg:   teeb
#   tud:   tehtud

# All forms
sonaveeb -all puu

# JSON output
sonaveeb -json puu
```
