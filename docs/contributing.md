# Contributing to oshiv

For detailed guidelines on submitting Enhancement Proposals and Pull Requests, please see the main [CONTRIBUTING.md](../CONTRIBUTING.md) file.

## Development

### Build

To build the project locally:

```
make build
```

### Test and Push

Test and validate your changes, push to your fork, and make a Pull Request:

1. Make your changes
2. Test locally
3. Push to your fork
4. Create a Pull Request following the guidelines in [CONTRIBUTING.md](../CONTRIBUTING.md)

### Release Process

After your PR has been reviewed and merged, the maintainers will perform the release process:

```
make release
```

Push version tag:

```
git tag -a <VERSION> -m '<COMMENTS>'
git push origin <VERSION>
```

Verify the [releaser](https://github.com/cnopslabs/oshiv/actions/workflows/releaser.yml) job completes successfully.