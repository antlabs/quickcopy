all:
	go build -o quickcopy ./cmd/quickcopy/quickcopy.go

intel-mac:
	GOARCH=amd64 go build -o quickcopy.intel ./cmd/quickcopy/quickcopy.go

test:
	./quickcopy
	go test ./...