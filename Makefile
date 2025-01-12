# check OS and set the correct path separator and commands
ifeq ($(OS),Windows_NT)
	FIXPATH = $(subst /,\,$1)
	RM = del /Q
	MKDIR = mkdir
else
	FIXPATH = $1
	RM = rm -rf
	MKDIR = mkdir -p
endif

GOEXE:=$(shell go env GOEXE)
GOOS:=$(shell go env GOOS)
GOARCH:=$(shell go env GOARCH)

NAME:=ged-shovel
OUTDIR:=dist
LONGNAME=$(NAME)-$(GOOS)-$(GOARCH)
BINNAME:=$(call FIXPATH,dist/$(LONGNAME)$(GOEXE))

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

$(OUTDIR):
	$(MKDIR) $@

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


.PHONY: test
test:
	go test ./...


.PHONY: build
build: $(OUTDIR)
	go build -o $(BINNAME) $(call FIXPATH,./cmd/$(NAME)/)
	@echo "Built $(BINNAME)"


.PHONY: release
release: clean $(OUTDIR) build release_$(GOOS)


.PHONY: release_linux
release_linux:
	cd $(OUTDIR) ; tar -czf ../$(LONGNAME).tar.gz *


.PHONY: release_windows
release_windows:
	zip -j $(LONGNAME).zip  $(OUTDIR)/*


.PHONY: run
run:
	go run $(call FIXPATH,./cmd/$(NAME)/) --metrics ":2112"


.PHONY: image
image:
	docker build -t $(NAME) .


.PHONY: no-dirty
no-dirty:
	git diff --exit-code


.PHONY: clean
clean:
	$(RM) $(call FIXPATH,$(OUTDIR))
