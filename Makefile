benchmark:
	go test -v ./... -run=^$ -bench=.

.PHONY: benchmark