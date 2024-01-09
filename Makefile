.PHONY: docker-dev
docker-dev:
	cd docker/compose ;\
	docker compose up -d
	docker ps

.PHONY: docker-dev-clean
docker-dev-clean:
	cd docker/compose ;\
	docker compose down -v

.PHONY: goimports
goimports:
	goimports -w  .

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: build
build:
	go build -gcflags "-m" -race -o ./cmd/userapi/userapi ./cmd/userapi/*.go

.PHONY: statictest
statictest:
	go vet -vettool=$$(which statictest) ./...

test-nginx:
	docker run --rm -it --entrypoint nginx -v ./docker/compose/nginx.conf:/etc/nginx/nginx.conf nginx:1.25.3-bookworm -t
