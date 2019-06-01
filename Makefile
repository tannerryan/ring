test:
	go test -v ./...

coverage:
	go test -covermode=count -coverprofile=count.out ./...
	go tool cover -func=count.out
	rm count.out

codecov:
	./go.test.sh

.SILENT:
