test:
	go clean -cache
	go test -race -v -coverprofile=c.out ./pkg/scsi ./pkg/multipath .

check:
	gofmt -w ./.
	golint ./...
	go vet

gocover:
	go tool cover -html=c.out
