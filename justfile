tidy: gen fmt
    go mod tidy

test:
    CGO_ENABLED=1 go test -count=1 -failfast -v ./...

test-race:
    CGO_ENABLED=1 go test -count=1 -v -race ./...

fmt:
    go tool fmt .

dep:
    go mod tidy

update:
    go get -u ./...

gen:
    go generate -v ./...
    go generate -v ./testdata/model/...

serve-dbs:
    cd hack && docker compose up -d
