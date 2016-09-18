.PHONY: all debug clean

all: clean
	go test -v ./... && go build -gcflags '$(FLAGS)' -v ./...

debug:
	$(MAKE) FLAGS="-N -l $(FLAGS)"

clean:
	go clean ./...