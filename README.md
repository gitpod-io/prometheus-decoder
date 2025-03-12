# prometheus-decoder

A simplistic tool to turn Prometheus metric records as persisted by Gitpod Dedicated into a readable form.

## Usage

```
$ ./prometheus-decoder
Error: Input file is required
Usage of ./prometheus-decoder:
  -human-time
        Show human-readable timestamps in output
  -input string
        Input file containing PrometheusRecord entries (one per line)
  -output string
        Output file for JSON results (default: stdout)
  -pretty
        Enable pretty-printing of JSON output (default true)
  -version
        Show version information and exit
```

## Installation

### Download from GitHub Releases

You can download the latest pre-built binary for your platform from the [GitHub Releases page](https://github.com/gitpod-io/prometheus-decoder/releases).

### Build from Source

```bash
git clone https://github.com/gitpod-io/prometheus-decoder.git
cd prometheus-decoder
go build
```

## Development

This project uses GitHub Actions to automatically build and publish releases when a new tag is pushed to the repository.

### Creating a New Release

1. Update the version number in `main.go`
2. Commit your changes
3. Tag the commit with a version number:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
4. GitHub Actions will automatically build binaries for multiple platforms and create a new release

### Version History

- v0.1.0 - Initial release
