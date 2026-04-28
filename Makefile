BUILD_FLAGS=-gcflags="all=-N -l" -trimpath -mod=readonly -modcacherw

.PHONY: build lint test install check run

build:
	go build $(BUILD_FLAGS) ./...

run:
	go run ./cmd/gitmd

install:
	go install $(BUILD_FLAGS) ./cmd/gitmd

lint:
	golangci-lint run --timeout 5m

test:
	go test -v -cover ./...

check:
	go mod tidy
	git diff --exit-code
	$(MAKE) lint
	$(MAKE) test
