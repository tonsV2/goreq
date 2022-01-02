tag ?= latest
version ?= $(shell yq e '.version' helm/Chart.yaml)
clean-cmd = docker compose down --remove-orphans --volumes

binary:
	go build -o goreq -ldflags "-s -w" ./cmd/goreq

install:
	go install github.com/tonsV2/goreq/cmd/goreq@latest

docker-image:
	IMAGE_TAG=$(tag) docker compose build prod

push-docker-image:
	IMAGE_TAG=$(tag) docker compose push prod

dev:
	docker compose up --build dev

test: clean
	docker compose run --no-deps test
	$(clean-cmd)

dev-test: clean
	docker compose run --no-deps dev-test
	$(clean-cmd)

clean:
	$(clean-cmd)
	go clean

di:
	wire gen ./internal/di

.PHONY: binary docker-image push-docker-image dev test dev-test
