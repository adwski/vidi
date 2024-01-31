.PHONY: mock
mock:
	find . -type f -name "mock_*" -exec rm -rf {} +
	mockery

.PHONY: docker-dev
docker-dev:
	cd docker/compose ;\
	docker compose up -d
	docker ps

docker-infra:
	cd docker/compose ;\
	docker compose -f docker-compose.infra.yml up -d
	docker ps

docker-infra-clean:
	cd docker/compose ;\
	docker compose -f docker-compose.infra.yml down -v
	docker ps

docker-dev-build:
	cd docker/compose ;\
	docker compose up -d --build
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

.PHONY: unittests
unittests:
	go test ./... -v -count=1 -cover -coverpkg=./... -coverprofile=profile.cov ./...
	go tool cover -func profile.cov

.PHONY: cover
cover:
	go tool cover -html profile.cov -o coverage.html


.PHONY: test-all
test-all: docker-infra
	go test -v -count=1 -cover -coverpkg=./... -coverprofile=profile.cov --tags e2e ./...
	go tool cover -func profile.cov
	$(MAKE) cover
	$(MAKE) docker-infra-clean
