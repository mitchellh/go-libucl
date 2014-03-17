all: vendor/libucl/libucl.a
	go build

clean:
	rm -rf vendor/
	rm -f libucl.dll

test: vendor/libucl/libucl.a
	go test

vendor/libucl/libucl.a:
	./scripts/build_libucl.sh

.PHONY: all clean test
