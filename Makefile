.PHONY: test

fmt:
	./script/fmt.sh

cover:
	./script/cover.sh
	go tool cover -html=coverage.txt

test:
	./script/test.sh