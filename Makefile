.PHONY: all debug clean test bench benchmem lines

all: clean
	go test -v ./... && go build -gcflags '$(FLAGS)' -v ./...

debug:
	$(MAKE) FLAGS="-N -l $(FLAGS)"

clean:
	go clean ./...

test:
	go test -v ./...

bench:
	go test -test.bench=.* ./... | grep "^Benchmark"

bench-mem:
	go test -test.bench=.* -benchmem ./... | grep "^Benchmark"

lines:
	find ./ -name "*.go" | xargs cat | wc
