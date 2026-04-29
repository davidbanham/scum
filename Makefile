.PHONY: test
test:
	set -o pipefail; go test -short `go list ./... | grep -v /vendor/` | grep -v NOTICE
