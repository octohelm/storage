module github.com/octohelm/storage

go 1.25.5

tool (
	github.com/octohelm/storage/internal/cmd/fmt
	github.com/octohelm/storage/internal/cmd/gen
)

// +gengo:import:group=0_controlled
require (
	github.com/octohelm/enumeration v0.0.0-20251117072411-c5ede10316bf
	github.com/octohelm/gengo v0.0.0-20251223082640-0fa9a703560f
	github.com/octohelm/x v0.0.0-20251028032356-02d7b8d1c824
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-json-experiment/json v0.0.0-20251027170946-4849db3c2f7e
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.8.0
	golang.org/x/sync v0.19.0
	modernc.org/sqlite v1.43.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20251219203646-944ab1f22d93 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	modernc.org/libc v1.67.4 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
)
