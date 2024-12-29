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

DIRS = dist

.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v


.PHONY: run
run: 
	go run $(call FIXPATH,./cmd/enp/) --metrics ":2112"


.PHONY: test
test:
	go test -v ./...


.PHONY: build
build: $(DIRS)
	cd dist & go build $(call FIXPATH,../cmd/enp/)


$(DIRS):
	$(MKDIR) $@