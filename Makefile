.PHONY: build test vet check vuln dist brand ui-test

brand:
	python3 scripts/brand_assets.py

ui-test:
	npm test

build:
	go build -trimpath -o dexrove .

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
