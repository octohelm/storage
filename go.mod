module github.com/octohelm/storage

go 1.25.4

tool (
	github.com/octohelm/storage/internal/cmd/fmt
	github.com/octohelm/storage/internal/cmd/gen
)

// +gengo:import:group=0_controlled
require (
	github.com/octohelm/enumeration v0.0.0-20250115091102-ddf41951d733
	github.com/octohelm/gengo v0.0.0-20251031090256-3806c0e4d0f8
	github.com/octohelm/x v0.0.0-20251028032356-02d7b8d1c824
)

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/duckdb/duckdb-go/v2 v2.5.1
	github.com/go-json-experiment/json v0.0.0-20251027170946-4849db3c2f7e
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	golang.org/x/sync v0.18.0
	modernc.org/sqlite v1.40.0
)

require (
	github.com/apache/arrow-go/v18 v18.4.1 // indirect
	github.com/duckdb/duckdb-go-bindings v0.1.22 // indirect
	github.com/duckdb/duckdb-go-bindings/darwin-amd64 v0.1.22 // indirect
	github.com/duckdb/duckdb-go-bindings/darwin-arm64 v0.1.22 // indirect
	github.com/duckdb/duckdb-go-bindings/linux-amd64 v0.1.22 // indirect
	github.com/duckdb/duckdb-go-bindings/linux-arm64 v0.1.22 // indirect
	github.com/duckdb/duckdb-go-bindings/windows-amd64 v0.1.22 // indirect
	github.com/duckdb/duckdb-go/arrowmapping v0.0.24 // indirect
	github.com/duckdb/duckdb-go/mapping v0.0.24 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/telemetry v0.0.0-20251008203120-078029d740a8 // indirect
	golang.org/x/text v0.30.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	modernc.org/libc v1.66.10 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
)
