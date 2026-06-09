# infernum

Benchmark your local Ollama models and compare performance across hardware.

## Install

### From source

```bash
git clone https://github.com/joelhelbling/infernum.git
cd infernum
make build
```

### Homebrew (coming soon)

```bash
brew install joelhelbling/tap/infernum
```

## Usage

### Run benchmarks

```bash
infernum run --models llama3:8b,mistral:7b
```

Runs the default benchmark suite against the specified models, publishes results, and prints a link to view them.

### Compare hardware for a model

```bash
infernum compare --model llama3:8b
```

### Compare models on hardware

```bash
infernum compare --hardware <config-id>
```

### Filter comparisons

```bash
infernum compare --model llama3:8b --gpu "RTX 4090" --ram-min 32
```

### View a specific run

```bash
infernum results <run-id>
```

### List benchmark suites

```bash
infernum suites
```

### JSON output (for agentic use)

All commands support `--format json` for structured output:

```bash
infernum compare --model llama3:8b --format json
```

## Configuration

Config file: `~/.config/infernum/config.yaml`

```yaml
api_base_url: https://bench.ollama.example.com
```

## Building from source

```bash
make build    # build binary
make test     # run unit tests
```
