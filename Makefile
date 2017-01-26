fmt:
	go fmt ./...

cover:
	./code-coverage.sh
	go tool cover -html=coverage.txt