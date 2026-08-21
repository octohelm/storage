module github.com/octohelm/storage

go 1.27.0

tool (
	github.com/octohelm/storage/tool/internal/cmd/fmt
	github.com/octohelm/storage/tool/internal/cmd/gen
	github.com/octohelm/storage/tool/internal/cmd/skills-install
)

// +gengo:import:group=0_controlled
require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.10.0
	// +skill:enumeration-guideline
	github.com/octohelm/enumeration v0.0.0-20260508105338-2e799c70cf82
	// +skill:gengo-guideline
	github.com/octohelm/gengo v0.0.0-20260821034500-ef88841eb5e4
	// +skill:testing-guideline
	github.com/octohelm/x v0.0.0-20260821032215-38f48df07f8c
	golang.org/x/sync v0.22.0
	modernc.org/sqlite v1.57.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	modernc.org/libc v1.75.4 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	mvdan.cc/gofumpt v0.11.0 // indirect
)
