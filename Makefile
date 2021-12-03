tag ?= latest
version ?= $(shell yq e '.version' helm/Chart.yaml)
clean-cmd = docker compose down --remove-orphans --volumes

binary:
	go build -o goreq -ldflags "-s -w" ./cmd/goreq

docker-image:
	IMAGE_TAG=$(tag) docker compose build prod

push-docker-image:
	IMAGE_TAG=$(tag) docker compose push prod

dev:
	docker compose up --build dev database redis

test: clean
	docker compose run --no-deps test
	$(clean-cmd)

dev-test: clean
	docker compose run --no-deps dev-test
	$(clean-cmd)

clean:
	$(clean-cmd)
	go clean

swagger-docs:
	swag init --parseDependency --parseDepth 2 -g ./cmd/goreq/main.go --output swagger/docs

swagger: swagger-docs

di:
	wire gen ./internal/di

.PHONY: binary docker-image push-docker-image dev test dev-test
