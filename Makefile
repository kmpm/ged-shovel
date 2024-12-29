# check OS and set the correct path separator and commands
ifeq ($(OS),Windows_NT)
	FIXPATH = $(subst /,\,$1)
	RM = del /Q
else
	FIXPATH = $1
	RM = rm -f
endif


tidy:
	go fmt ./...
	go mod tidy -v

run: 
	go run $(call FIXPATH,./cmd/enp/) --metrics ":2112"


test:
	go test -v ./...