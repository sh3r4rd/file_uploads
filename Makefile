.PHONY: build deploy test clean fmt lint init validate

build:
	./scripts/build.sh

deploy:
	./scripts/deploy.sh

test:
	cd lambdas && go test ./...

clean:
	rm -rf lambdas/bin/

fmt:
	cd lambdas && go fmt ./...
	cd infra && terraform fmt

lint:
	cd lambdas && go vet ./...

init:
	cd infra && terraform init

validate:
	cd infra && terraform validate
