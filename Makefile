GOPATH= $(shell go env GOPATH)

EXE = 
ifdef OS
	EXE = .exe
endif

.PHONY: test
test:
	go test -race ./...

.PHONY: lint
lint:
	go vet ./...
	GO111MODULE=on go get honnef.co/go/tools/cmd/staticcheck@2020.1.3
	$(GOPATH)/bin/staticcheck -go 1.14 ./...

# Exclude rules from security check:
# G104 (CWE-703): Errors unhandled.
.PHONY: seccheck
seccheck:
	go vet ./...
	GO111MODULE=on go get github.com/securego/gosec/v2/cmd/gosec
	$(GOPATH)/bin/gosec -exclude=G104 ./...
