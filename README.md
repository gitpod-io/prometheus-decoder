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
```