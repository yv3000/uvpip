module github.com/yv3000/uvpip

go 1.26.0

// Dev-only build and CI audit tools (never linked into uvpip binary)
tool (
	github.com/rhysd/actionlint/cmd/actionlint
	golang.org/x/vuln/cmd/govulncheck
)

// Transitive dependencies required exclusively by dev-tools (actionlint and govulncheck); 0 runtime dependencies linked into uvpip binary
require (
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect; transitive of actionlint (glob matcher)
	github.com/fatih/color v1.19.0 // indirect; transitive of actionlint (terminal colors)
	github.com/mattn/go-colorable v0.1.14 // indirect; transitive of actionlint (windows console colors)
	github.com/mattn/go-isatty v0.0.20 // indirect; transitive of actionlint (tty detection)
	github.com/mattn/go-runewidth v0.0.21 // indirect; transitive of actionlint (east asian width)
	github.com/mattn/go-shellwords v1.0.12 // indirect; transitive of actionlint (shell line parser per go mod why)
	github.com/rhysd/actionlint v1.7.12 // tool binary implementation
	github.com/robfig/cron/v3 v3.0.1 // indirect; transitive of actionlint (cron expression parser per go mod why)
	golang.org/x/mod v0.41.0 // indirect; transitive of govulncheck/actionlint
	golang.org/x/sync v0.23.0 // indirect; transitive of govulncheck (errgroup)
	golang.org/x/sys v0.48.0 // indirect; transitive of actionlint/govulncheck
	golang.org/x/telemetry v0.0.0-20260908163034-4bcc4b2ee518 // indirect; transitive of govulncheck
	golang.org/x/tools v0.50.0 // indirect; transitive of govulncheck
	golang.org/x/vuln v1.8.0 // tool binary implementation
)

require (
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.3 // indirect
)
