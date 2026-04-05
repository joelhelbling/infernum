# ollama-bench

Benchmark your local Ollama models and compare performance across hardware.

## Install

### From source

```bash
git clone https://github.com/joelhelbling/ollama-bench.git
cd ollama-bench
make build
```

### Homebrew (coming soon)

```bash
brew install joelhelbling/tap/ollama-bench
```

## Usage

### Run benchmarks

```bash
ollama-bench run --models llama3:8b,mistral:7b
```

Runs the default benchmark suite against the specified models, publishes results, and prints a link to view them.

### Compare hardware for a model

```bash
ollama-bench compare --model llama3:8b
```

### Compare models on hardware

```bash
ollama-bench compare --hardware <config-id>
```

### Filter comparisons

```bash
ollama-bench compare --model llama3:8b --gpu "RTX 4090" --ram-min 32
```

### View a specific run

```bash
ollama-bench results <run-id>
```

### List benchmark suites

```bash
ollama-bench suites
```

### JSON output (for agentic use)

All commands support `--format json` for structured output:

```bash
ollama-bench compare --model llama3:8b --format json
```

## Configuration

Config file: `~/.config/ollama-bench/config.yaml`

```yaml
api_base_url: https://bench.ollama.example.com
```

## Building from source

```bash
make build    # build binary
make test     # run unit tests
```
