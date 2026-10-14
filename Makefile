.PHONY: build test vet check vuln dist

build:
	go build -trimpath -o pokedex .

test:
	go test -race ./...

vet:
	go vet ./...

check: test vet
	@test -z "$$(gofmt -l $$(git ls-files '*.go'))"
	git diff --check

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...

dist:
	python3 scripts/release.py $(VERSION)
