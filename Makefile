GO?=go

test:
	@$(GO) test -v -race -cover

coverage:
	@$(GO) test -race -covermode=atomic -coverprofile=cover.out

.PHONY: test coverage
