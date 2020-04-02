.PHONY: test testdel example
testdel:
	go test ./... -coverprofile=cover.out -v
	go tool cover -func=cover.out
	rm -f cover.out
test:
	go test ./... -coverprofile=cover.out
	go tool cover -func=cover.out
example:
	cd example && rm -Rf data && go generate && go run *.go && cd -
