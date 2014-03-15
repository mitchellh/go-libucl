all: vendor/libucl/libucl.a
	go build

test: vendor/libucl/libucl.a
	go test

vendor/libucl/libucl.a:
	./scripts/build_libucl.sh

.PHONY: all test
