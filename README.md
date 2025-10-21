# Go Application Template

🧩 Your App — Go Project Template (2025)

A clean, production-grade Go project template with:
* 	✅ Standard Go module layout (cmd/, internal/, pkg/)
* 	🧹 Strict linting via golangci-lint
* 	🪄 Auto-formatting and pre-commit checks
* 	⚡ Ready-to-run Makefile (build, test, lint, etc.)
* 	🔒 Modern Go 1.23+ practices (slog, go:embed, etc.)
* 	🧪 CI/CD-friendly (GitHub Actions / pre-commit)

```shell
your-app/
├─ cmd/
│  └─ your-app/
│     └─ main.go
├─ internal/
│  └─ app/
│     └─ app.go
├─ go.mod
├─ go.sum                 # (generated)
├─ Makefile
├─ .golangci.yml
├─ .pre-commit-config.yaml
├─ .editorconfig
├─ .gitignore
└─ .hooks/
   └─ pre-commit          # optional plain Git hook (if you don't use pre-commit framework)
```

⚙️ Quick Start
```shell
# Clone
git clone https://github.com/yourname/your-app
cd your-app

# Initialize modules
go mod tidy

# (optional) install pre-commit framework
brew install pre-commit || pipx install pre-commit
pre-commit install

# Run checks (fmt, vet, lint, test)
make verify

# Build binary
make build
./bin/your-app
```

🧪 Development Commands
```shell
| Command       | Description                         |
|----------------|-------------------------------------|
| `make build`   | Compile binary to `bin/`            |
| `make run`     | Run app directly                    |
| `make verify`  | Run tidy, fmt, vet, lint, test      |
| `make fmt`     | Run `go fmt` + `gofumpt`            |
| `make lint`    | Run `golangci-lint`                 |
| `make test`    | Run unit tests                      |
| `make cover`   | Run coverage report                 |
| `make vuln`    | Run vulnerability scanner           |
| `make clean`   | Clean build artifacts               |

```

🧰 Tooling Setup

🪶 Linter

Configured in .golangci.yml, with modern rules:
* 	gofumpt, goimports, staticcheck, revive, copyloopvar, etc.
* 	Opinionated formatting and style enforcement
* 	make lint runs everything consistently across dev & CI

🔍 Pre-commit Hooks

Configured via .pre-commit-config.yaml:
* 	Runs go fmt, go vet, go mod tidy, and golangci-lint --fast
* 	Keeps code clean before you even push

Install & enable:
```shell
brew install pre-commit
pre-commit install
```

🧾 Editor Config

.editorconfig ensures consistent whitespace, line endings, and tabs across editors.

⸻

🧠 Design Philosophy

This template follows idiomatic Go project layout:
* 	cmd/ — entrypoints (binaries)
* 	internal/ — private packages (non-exported)
* 	pkg/ (optional) — shared packages (if needed)
* 	Makefile — one command for local or CI
* 	slog for structured logging (Go 1.21+)
* 	Minimal dependencies: focuses on maintainability

⸻

