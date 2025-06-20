# CLI Testing

This directory contains tests for the CLI, primarily using golden file testing.

## Golden File Testing

Golden file testing compares the output of CLI commands against expected "golden" files stored in version control.

### Running Tests

To run all tests:

```bash
cd tests
go test -v
```

### Updating Golden Files

When you intentionally change CLI behavior or output format:

1. Run the golden file generation script:

```bash
./scripts/generate_golden.sh
```

2. Review the changes to ensure they match your intended modifications
3. Commit the updated golden files along with your code changes

### Adding New Tests

To add tests for a new command:

1. Create a new directory under `golden/` for your command
2. Add the command to the `commands` array in `scripts/generate_golden.sh`
3. Run `./scripts/generate_golden.sh` to generate the golden files
4. Review the output and commit the new files

### Troubleshooting

If tests fail with line ending differences, check:

- Git configuration for line endings
- Text editor settings
- OS-specific differences in output generation

For dynamic content (like timestamps), consider modifying the test to handle these special cases.
