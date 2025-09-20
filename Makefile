TOOLS_SHELL="./hack/tools.sh"

.PHONY: test
test:
	@${TOOLS_SHELL} test
	@echo "go test finished"



.PHONY: vet
vet:
	@${TOOLS_SHELL} vet
	@echo "vet check finished"

build:
	go build -o cwgo.exe .

install: build
	cp cwgo.exe C:\Users\yihui\go\bin