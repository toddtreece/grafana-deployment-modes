.PHONY: all
all: 
	@go run .

.PHONY: sender
sender:
	@go run . -target sender

.PHONY: server
server:
	@go run . -target server
