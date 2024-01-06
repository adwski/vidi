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
