ifeq ("$(origin V)", "command line")
  VERBOSE = $(V)
endif
ifneq ($(VERBOSE),1)
.SILENT:
endif

build:
	GOFLAGS=-trimpath gox $(if $(VERBOSE),-verbose,) \
		-os=linux -arch=arm64 -ldflags="-s -w" -tags lambda.norpc \
		-output="bin/{{.Dir}}/bootstrap" ./...

lint:
	-golangci-lint run $(if $(VERBOSE),-v,) ./...

test:
	gotest $(if $(VERBOSE),-v,) -cover ./...
