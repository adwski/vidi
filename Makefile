SHELL := /bin/bash

.PHONY: mock
mock:
	find . -type f -name "mock_*" -exec rm -rf {} +
	mockery

.PHONY: docker-dev
docker-dev:
	cd docker/compose ;\
	docker compose up -d

.PHONY: docker-infra
docker-infra:
	cd docker/compose ;\
	docker compose -f docker-compose.infra.yml -f docker-compose.nginx.yml up -d

.PHONY: docker-infra-clean
docker-infra-clean:
	cd docker/compose ;\
	docker compose -f docker-compose.infra.yml -f docker-compose.nginx.yml down -v

.PHONY: docker-dev-build
docker-dev-build:
	cd docker/compose ;\
	docker compose up -d --build

.PHONY: docker-dev-build-svc
docker-dev-build-svc:
	cd docker/compose ;\
	docker compose up -d --no-deps --build $$SVC

.PHONY: docker-dev-clean
docker-dev-clean:
	cd docker/compose ;\
	docker compose down -v

.PHONY: docker-ps
docker-ps:
	docker ps --format="table {{.Image}}\t{{.Status}}\t{{.Names}}"

.PHONY: goimports
goimports:
	goimports -w  .

.PHONY: lint
lint:
	golangci-lint run ./...

build-tui:
	goreleaser build --debug --clean

.PHONY: statictest
statictest:
	go vet -vettool=$$(which statictest) ./...

.PHONY: test-nginx
test-nginx:
	docker run --rm -it --entrypoint nginx \
		-v ./docker/compose/tls:/etc/nginx/tls \
		-v ./docker/compose/nginx.conf:/etc/nginx/nginx.conf \
		nginx:1.25.3-bookworm -t

.PHONY: unittests
unittests:
	go test -v -count=1 ./...

.PHONY: cover
cover:
	go tool cover -html profile.cov -o coverage.html

.PHONY: test-all
test-all: docker-infra
	go test -v -count=1 -cover -coverpkg=./... -coverprofile=profile.cov --tags e2e ./...
	go tool cover -func profile.cov
	$(MAKE) cover
	$(MAKE) docker-infra-clean

.PHONY: grpc
grpc:
	protoc --go_out=.  --go-grpc_out=.  internal/api/video/grpc/protobuf/user.proto
	protoc --go_out=.  --go-grpc_out=.  internal/api/video/grpc/protobuf/service.proto

.PHONY: tls
tls:
	openssl req -x509 -nodes -days 365 -newkey rsa:4096 \
	-keyout ./docker/compose/tls/key.pem -out ./docker/compose/tls/cert.pem \
	-subj "/C=RU/O=ViDi/OU=vidi/CN=vidi" \
		-addext "subjectAltName = DNS:localhost, IP:127.0.0.1, IP:::1"

	VIDI_CA=$$(base64 -i ./docker/compose/tls/cert.pem) && \
		tmp=$$(mktemp) &&  \
		cfg="./docker/compose/config/config.json"  && \
		jq --arg a "$$VIDI_CA" '.vidi_ca = $$a' $$cfg > $$tmp && \
		mv $$tmp $$cfg
