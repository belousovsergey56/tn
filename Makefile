.DEFAULT_GOAL := build

fmt:
	go fmt ./..
.PHONY:fmt

# lint: fmt
# 	go lint ./..
# .PHONY:lint

# vet: fmt
# 	go vet ./..
# .PHONY:vet

build: fmt
	go build -o tn main.go
.PHONY:build

run:
	go run main.go
.PHONY:run
