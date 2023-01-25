.PHONY: all
all: 
	@go run .

.PHONY: client
client:
	@go run . -target sender

.PHONY: server
server:
	@go run . -target server
