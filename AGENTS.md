# sonaveeb-cli

## Commands

Use the standard Make targets shared across these Go CLI repos:

- Format: `make fmt`
- Lint: `make lint`
- Test: `make test`
- Build: `make build`
- Check: `make check`
- Install: `make install`

`make build` formats first. `make check` runs lint, test, then build.

For full testing including integration tests (requires `EKILEX_API_KEY`):

```bash
go test -tags=integration
```

## Design Principles

Design for testability using "functional core, imperative shell": keep pure business logic separate from code that does IO.

## Versioning

Always bump the CLI version when making changes.
