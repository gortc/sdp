PROCS := $(shell nproc)
lint:
	@echo "linting on $(PROCS) cores"
	@gometalinter -e "\.String\(\).+gocyclo" \
		-e "_test.go.+(gocyclo|errcheck|dupl)" \
		-e "isZeroOrMore is a pure function but its return value is ignored" \
		-e "isOptional is a pure function but its return value is ignored" \
		-e "parameter result 0 \(bool\) is never used" \
		-e "parameter d always receives \"IN\"" \
		--enable-all \
		--enable="lll" --line-length=100 \
		--disable=gocyclo \
		--disable=gochecknoglobals \
		--deadline=300s \
		--dupl-threshold=70 \
		-j $(PROCS)
	@gocritic check-project .
	@echo "ok"
install:
	go get gortc.io/api
	go get -u github.com/go-critic/go-critic/...
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install --update
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

