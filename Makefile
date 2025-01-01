# check OS and set the correct path separator and commands
ifeq ($(OS),Windows_NT)
	FIXPATH = $(subst /,\,$1)
	RM = del /Q
	MKDIR = mkdir
else
	FIXPATH = $1
	RM = rm -f
	MKDIR = mkdir -p
endif

.PHONY: help
help:
	@echo "make tidy"
	@echo "       go fmt and go mod tidy"
	@echo "make run"
	@echo "       run the application"
	@echo "make test"
	@echo "       run the tests"
	@echo "make build"
	@echo "       build the application"
	@echo "make no-dirty"
	@echo "       check if there are any uncommitted changes"
	@echo "make help"
	@echo "       show this help message"


.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v


.PHONY: audit
audit:
	@echo "running checks"
	go mod verify
	go vet ./...
	go list -m all
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...


.PHONY: no-dirty
no-dirty:
	git diff --exit-code


.PHONY: build
build:
	go build $(call FIXPATH,./cmd/ged-shovel/)


.PHONY: test
test:
	go test ./...


.PHONY: image
image:
	docker build -t ged-shovel .

.PHONY: run
run:
	go run $(call FIXPATH,./cmd/ged-shovel/) --metrics ":2112"
