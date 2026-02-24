module github.com/octohelm/storage

go 1.26.0

tool (
	github.com/octohelm/storage/internal/cmd/fmt
	github.com/octohelm/storage/internal/cmd/gen
)

// +gengo:import:group=0_controlled
require (
	github.com/octohelm/enumeration v0.0.0-20260224023935-6eaef7930a8b
	github.com/octohelm/gengo v0.0.0-20260224022252-ec6c2fc2f701
	github.com/octohelm/x v0.0.0-20260224021741-724787200747
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-json-experiment/json v0.0.0-20260214004413-d219187c3433
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.8.0
	golang.org/x/sync v0.19.0
	modernc.org/sqlite v1.46.1
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
	golang.org/x/exp v0.0.0-20260218203240-3dfff04db8fa // indirect
	golang.org/x/mod v0.33.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	golang.org/x/tools v0.42.0 // indirect
	modernc.org/libc v1.68.0 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
)
