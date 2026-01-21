# sonaveeb-cli

## Testing

After any changes, run:

```bash
go build && go test
```

For full testing including integration tests (requires EKILEX_API_KEY):

```bash
go test -tags=integration
```

## Design Principles

Design for testability using "functional core, imperative shell": keep pure business logic separate from code that does IO.

## Versioning

Always bump the CLI version when making changes.
