BUILD_FLAGS=-gcflags="all=-N -l" -trimpath -mod=readonly -modcacherw

build:
	go build $(BUILD_FLAGS) ./...

lint:
	golangci-lint run --timeout 5m

test:
	go test -v -cover ./...

build:
	go build $(BUILD_FLAGS) ./...

install:
	go install go.uber.org/nilaway/cmd/nilaway@latest

check:
	go mod tidy
	git diff --exit-code
	make lint
	make test
