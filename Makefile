PROCS := $(shell nproc)
lint:
	@echo "linting on $(PROCS) cores"
	@gometalinter -e "\.String\(\).+gocyclo" \
		-e "_test.go.+(gocyclo|errcheck|dupl)" \
		--enable="lll" --line-length=80 \
		--enable="gofmt" \
                --disable=gocyclo \
		--deadline=300s \
                --dupl-threshold=70 \
		-j $(PROCS)
	@echo "ok"
install:
	go get -u sourcegraph.com/sqs/goreturns
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install --update
	go get -u github.com/cydev/go-fuzz/go-fuzz-build
	go get -u github.com/dvyukov/go-fuzz/go-fuzz
format:
	goimports -w .
profile:
	go tool pprof -alloc_space sdp.test mem.out
profile-cpu:
	go tool pprof sdp.test cpu.out
