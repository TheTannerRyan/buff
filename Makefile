test:
	GOCACHE=off go test -race -v buff_test.go

coverage:
	GOCACHE=off go test -covermode=count -coverprofile=count.out ./...
	go tool cover -func=count.out
	rm count.out

.SILENT:
