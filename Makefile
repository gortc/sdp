PROCS := $(shell nproc)
lint:
	@golangci-lint run
	@echo "ok"
install:
	go get gortc.io/api
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
format:
	goimports -w .
profile:
	go tool pprof -alloc_space sdp.test mem.out
profile-cpu:
	go tool pprof sdp.test cpu.out
check-api:
	api -c api/sdp1.txt github.com/gortc/sdp
test:
	@./go.test.sh
test-e2e:
	@cd e2e && ./test.sh
