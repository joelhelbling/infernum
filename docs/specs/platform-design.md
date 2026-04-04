# Ollama Bench Platform Design

Shared performance benchmarking of local LLM performance. A CLI runs benchmarks against Ollama models on the user's hardware and publishes results to a web service where anyone can compare performance across models and hardware configurations.

## Repositories

### `ollama-bench` (open source, Go)

CLI tool distributed as a signed self-contained binary. Users can `brew install` or clone the repo, review the code, and build it themselves. Open for review, closed for contribution.

### `ollama-bench-web` (closed source)

Go REST API + Next.js frontend. Stores benchmark data in PostgreSQL. Serves comparison reports for humans (browser) and agents (CLI/JSON).

### Shared code

The web backend imports Go packages from the CLI repo (data models, API request/response types). The CLI never imports from the web repo. Shared packages live in `pkg/` in the CLI repo; CLI-internal logic lives in `internal/`.

The shared surface should be kept small (data models, API types) to minimize the coupling. Changes to shared packages are breaking changes for the web backend.

## Deployment

Both the Go API and Next.js frontend deploy as separate containers on Coolify (Hetzner, Arm64 Ampere). PostgreSQL is provisioned via Coolify's managed database support.

Architecture follows Approach A: the Go API is the single source of truth. The Next.js frontend is a separate service that calls the Go API. Both the CLI and the frontend are first-class API clients.

Next.js uses server-side rendering for pages that need social previews (shared report links). At request time, Next.js fetches data from the Go API and renders full HTML with meta tags.

---

## Benchmark Suites

### MVP scope

One canonical suite, stored in the backend. No suite creation/editing UI. The suite is viewable on the web UI (read-only).

### Suite definition

A suite consists of:

- **Unique identifier** (UUID) and **slug** (e.g., "default")
- **Version** (string)
- **Name** and **description**
- **Prompts**, each with:
  - Name (e.g., "simple_math", "manufacturing_calc")
  - Category (e.g., "coding", "reasoning", "general")
  - Prompt text
  - Optional expected token length range (min/max, for detecting anomalous runs)
  - Sort order
- **Parameters:**
  - Runs per prompt (default: 3)
  - Timeout per run (seconds)
  - Cooldown between runs (seconds, to reduce thermal throttling effects)

### Models are not part of the suite

Models are specified at run time by the user, since available models change frequently.

### Suite versioning

Results are tagged with suite ID + version. Results from different suite versions are flagged as non-comparable. This allows prompts to evolve without invalidating historical comparisons.

### Future scope (deferred)

Multiple suites, user-created suites, suite moderation. When implemented, suites will need guards against malicious prompt content.

---

## CLI Design

### Commands

```
ollama-bench run --models <model1,model2> [--suite <id>]
```
Fetches suite from backend (defaults to `default`). Detects hardware automatically. Runs each model x prompt x runs_per_prompt with cooldown between runs. Publishes results to backend. Prints a link to view results on the web. Displays a brief summary table before exiting.

```
ollama-bench compare --model <model> [--format json|table]
```
Queries backend for all hardware configs that have results for this model. Displays summary stats (mean eval rate, stddev) per hardware config.

```
ollama-bench compare --hardware <config-id> [--format json|table]
```
Queries backend for all models that have results on this hardware config. Displays summary stats per model.

```
ollama-bench results <run-id> [--format json|table]
```
Fetches and displays a specific run's results.

```
ollama-bench suites [--format json|table]
```
Lists available suites from the backend.

```
ollama-bench suite <id> [--format json|table]
```
Shows suite detail (prompts, parameters).

```
ollama-bench publish [<local-results-file>]
```
Publishes pending local results that failed to upload.

### Output formats

- `table` (default): ASCII tables for human reading in terminal
- `json`: structured output for agentic AI consumption (Claude Code, MCP, etc.)

### Configuration

- Backend URL: `~/.config/ollama-bench/config.yaml` (defaults to production service)
- Anonymous token: `~/.config/ollama-bench/token` (auto-generated on first run, persists across runs to identify this machine)

### Hardware detection

Auto-detects OS, CPU model, CPU cores, RAM, GPU, VRAM. Cross-platform (Linux, macOS, Windows). Hardware specs are sent with results for grouping/comparison.

---

## Network and Offline Behavior

### Suite fetching

The CLI requires internet to fetch a suite on first run. Once fetched, the suite is cached locally at `~/.cache/ollama-bench/suites/<id>-<version>.json`. On subsequent runs, the CLI checks for a newer version; if the service is unreachable, it uses the cached version and prints a notice. If there is no cached suite and the service is unreachable, the CLI exits with an error.

### Publishing

If the service is unreachable after a benchmark run, results are saved to `~/.local/share/ollama-bench/pending/` as JSON files. The user can publish them later with `ollama-bench publish`.

### Reporting commands

`compare`, `results`, and `suites` commands require internet access. The data lives on the server; there is no offline fallback for these.

### Timeouts

- 10s for fetching suites
- 30s for publishing results
- Individual benchmark runs respect the suite's timeout parameter

### CLI versioning

The CLI sends its version in an `X-OllamaBench-Version` header. The backend can reject submissions from outdated CLI versions with a clear upgrade message. The API is versioned via path (`/api/v1/...`).

---

## Backend API

### Base path: `/api/v1`

All endpoints are public (no auth in MVP). All responses are JSON.

### Endpoints

```
GET  /suites                      List available suites
GET  /suites/:id                  Suite detail (prompts, parameters)

POST /results                     Publish benchmark results
                                  Accepts: run data + hardware info
                                  Returns: { run_id, token, url }

GET  /results/:run-id             Single run results
GET  /results/:run-id?token=x     Same, sets cookie to mark "my run"

GET  /compare                     Unified comparison endpoint
                                  Query params for any combination of filters
                                  Returns: available values per dimension + summary stats

GET  /hardware-configs            List distinct hardware configs with submission counts
GET  /models                      List models with submission counts
```

### Compare endpoint

`GET /api/v1/compare` accepts any combination of filters:

- `model` - filter by model name
- `gpu` - filter by GPU name
- `cpu` - filter by CPU model
- `ram_min`, `ram_max` - filter by RAM range
- `vram_min`, `vram_max` - filter by VRAM range
- `os` - filter by OS name
- `arch` - filter by architecture
- `suite_id` - filter by suite (defaults to "default")
- `suite_version` - filter by suite version (defaults to latest)
- `expand=runs` - include individual run data alongside summary stats

Returns:
- Available values for each dimension given current filters (so the UI can populate filter options)
- Summary stats (mean eval rate, mean prompt eval rate, stddev, run count) for each matching group
- Individual runs when `expand=runs` is set

### Anonymous token flow

1. CLI sends first submission with hardware specs, no token
2. Backend generates a token, associates it with the hardware fingerprint, returns it
3. CLI persists the token locally
4. On subsequent submissions, CLI sends the token; backend associates the run with the same anonymous identity
5. The web link includes the token as a query param; visiting it sets a cookie so the UI can distinguish "my runs"

### Future: account claiming

Users create an account on the website and claim one or more anonymous tokens (one per machine). This associates all past and future submissions from those machines with their account. Deferred from MVP.

---

## Database Schema

### Tables

```sql
suites
  id                          UUID PRIMARY KEY
  slug                        VARCHAR UNIQUE
  name                        VARCHAR
  description                 TEXT
  version                     VARCHAR
  created_at                  TIMESTAMP
  updated_at                  TIMESTAMP

suite_prompts
  id                          UUID PRIMARY KEY
  suite_id                    UUID FK -> suites
  name                        VARCHAR
  category                    VARCHAR
  prompt_text                 TEXT
  expected_token_range_min    INT (nullable)
  expected_token_range_max    INT (nullable)
  sort_order                  INT

hardware_configs
  id                          UUID PRIMARY KEY
  os_name                     VARCHAR
  os_version                  VARCHAR
  architecture                VARCHAR
  cpu_model                   VARCHAR
  cpu_cores                   INT
  ram_gb                      FLOAT
  gpu_name                    VARCHAR (nullable)
  vram_gb                     FLOAT (nullable)
  fingerprint                 VARCHAR UNIQUE

runs
  id                          UUID PRIMARY KEY
  anonymous_token             VARCHAR
  suite_id                    UUID FK -> suites
  suite_version               VARCHAR
  hardware_id                 UUID FK -> hardware_configs
  model_name                  VARCHAR
  started_at                  TIMESTAMP
  completed_at                TIMESTAMP
  created_at                  TIMESTAMP

run_results
  id                          UUID PRIMARY KEY
  run_id                      UUID FK -> runs
  prompt_id                   UUID FK -> suite_prompts
  run_number                  INT
  total_duration_s            FLOAT
  load_duration_s             FLOAT
  prompt_eval_count           INT
  prompt_eval_duration_s      FLOAT
  prompt_eval_rate            FLOAT
  eval_count                  INT
  eval_duration_s             FLOAT
  eval_rate                   FLOAT
  success                     BOOLEAN
  error_message               TEXT (nullable)

anonymous_tokens
  token                       VARCHAR PRIMARY KEY
  hardware_id                 UUID FK -> hardware_configs
  created_at                  TIMESTAMP
  claimed_by                  UUID FK -> accounts (nullable, future)
```

### Indexes

**runs:**
- `(model_name)`
- `(hardware_id)`
- `(suite_id, suite_version)`

**run_results:**
- `(run_id)`

**hardware_configs:**
- `(fingerprint)` - deduplication
- `(gpu_name)` - filter by GPU
- `(cpu_model)` - filter by CPU
- `(ram_gb)` - filter by RAM
- `(vram_gb)` - filter by VRAM
- `(os_name)` - filter by OS
- `(architecture)` - filter by architecture

### Hardware deduplication

When the CLI submits results, the backend computes a fingerprint (hash of key hardware fields). If a matching config exists in `hardware_configs`, it reuses it. This powers the "which hardware configs have results for model X?" queries via joins on `hardware_id`.

### Core comparison query pattern

```sql
SELECT hc.*, AVG(rr.eval_rate), STDDEV(rr.eval_rate), COUNT(DISTINCT r.id)
FROM runs r
JOIN hardware_configs hc ON r.hardware_id = hc.id
JOIN run_results rr ON rr.run_id = r.id
WHERE r.model_name = ?
  AND r.suite_id = ? AND r.suite_version = ?
  AND rr.success = true
GROUP BY hc.id
```

---

## Web Frontend

### Tech stack

Next.js (TypeScript) with React. TanStack Query for data fetching/caching. Recharts or Nivo for charting. URL search params as source of truth for report state.

### Pages

#### Home / Landing

- Brief explanation of the service
- Link to install the CLI
- Link to view the default suite
- Quick stats (total runs, unique hardware configs, unique models)
- Two calls to action:
  - **"Compare hardware for a model"**: text/dropdown search for a model, then lands on the comparison page with that model pre-selected
  - **"Compare models on hardware"**: hardware selector, then lands on the comparison page with those hardware filters pre-selected

#### Suite Detail (`/suite/:slug`)

- Suite name, description, version
- List of prompts with name, category, prompt text
- Parameters (runs per prompt, timeout, cooldown)

#### Comparison Page (`/compare`)

One unified page, not two modes. All dimensions are filters:

- **Model** (from `runs.model_name`)
- **GPU, CPU, RAM, VRAM, OS, Architecture** (from `hardware_configs`)

Selecting a value in any dimension narrows the available options in all other dimensions to only combinations with actual submitted data.

**Flow:**
1. User selects/filters dimensions (via dropdowns, sliders, search)
2. UI shows matching hardware configs or models (depending on what's been filtered) with run counts
3. User checkboxes the specific items they want to compare
4. Report displays side-by-side summary stats (mean eval rate, mean prompt eval rate, stddev, run count)
5. Each group is expandable to show individual runs

**Entry from landing page:**
- "Compare hardware for a model" -> model pre-selected, hardware dimensions open
- "Compare models on hardware" -> hardware filters pre-selected, model dimension open

**Entry from Run Detail:**
- "Compare with other hardware" -> model pre-selected, user's own hardware config pre-checked, all other hardware unfiltered so user can add configs to compare

**Shareable URLs:** all filter selections encoded in search params.

#### Run Detail (`/results/:run-id`)

- Linked from CLI after publishing
- Shows all results: model, hardware, per-prompt stats
- If visited with anonymous token in URL, sets a cookie to mark "my runs"
- Link to jump to comparison page pre-filtered to this model/hardware

### Charting

- Bar charts for side-by-side eval rate comparisons
- Possible scatter plots for prompt eval rate vs generation eval rate

---

## Performance Metrics

The system tracks these metrics per benchmark execution. The focus is on raw token throughput (prompt processing and generation), not response quality.

| Metric | Description |
|---|---|
| `total_duration_s` | Full request time |
| `load_duration_s` | Model load time |
| `prompt_eval_count` | Prompt token count |
| `prompt_eval_duration_s` | Prompt processing time |
| `prompt_eval_rate` | Prompt processing speed (tokens/s) |
| `eval_count` | Generated token count |
| `eval_duration_s` | Generation time |
| `eval_rate` | Generation speed (tokens/s) |

---

## Testing Strategy

### CLI (Go)

- Unit tests for hardware detection, stats parsing, suite caching logic
- Integration tests that mock the HTTP API (test full `run` flow against a fake server)
- No tests against real Ollama (slow, environment-dependent). Test parsing of Ollama output, not Ollama itself.

### Backend (Go)

- Unit tests for data validation, fingerprint generation, aggregation query building
- Integration tests against a real Postgres instance (Docker in CI) for compare/filter queries
- API endpoint tests (HTTP-level) for full request/response cycle

### Frontend (Next.js)

- Component tests for filter/comparison UI (React Testing Library)
- E2E tests for the critical flow: landing -> select model -> filter hardware -> view comparison (Playwright)

### Cross-system

Smoke test in CI: CLI publishes results to a test backend (Docker Compose with Postgres + Go backend + CLI binary), then the compare endpoint returns them correctly.

---

## Deferred Work

- **Account system:** user registration, login, claiming anonymous tokens to an account
- **Multiple suites:** suite creation/editing UI, suite moderation, malicious content guards
- **Response evaluation:** criteria for evaluating response correctness/quality beyond raw throughput
- **MCP server:** thin wrapper over CLI commands for native agentic integration
- **Social preview images:** dynamic OG images for shared report links

---

## Implementation Notes

### Signed binary distribution

The CLI must ship with signed checksums (Sigstore/cosign) so users can verify that `brew install` binaries match the source. This is a requirement, not deferred. Trust is a core design principle of the open-source CLI.

### Hardware fingerprint computation

The fingerprint is a SHA-256 hash of the concatenation of normalized, lowercased values: `os_name + os_version + architecture + cpu_model + cpu_cores + ram_gb (rounded to nearest GB) + gpu_name + vram_gb (rounded to nearest GB)`. Rounding RAM/VRAM avoids false mismatches from OS reporting variance. The fingerprint is opaque to the user; hardware details are always displayed from the stored fields, not derived from the fingerprint.

### Suite version format

Suite versions use semver (e.g., "1.0.0"). Patch bumps for typo fixes or metadata changes (still comparable). Minor bumps for adding prompts (results from older versions are flagged but can still be displayed). Major bumps for removing/changing prompts (non-comparable).
