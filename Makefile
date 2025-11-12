.PHONY: help
help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ==============================================================================
# Main

.PHONY: run
run: ## run application
	go run ./main.go

.PHONY: build
build: ## build the project
	@make deps-tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./main.go

.PHONY: test
COVERING_PKG?=$$(go list ./... | grep -v  mock$ | grep -v script)
test:   ## run tests
	CGO_ENABLED=0 go test $(COVERING_PKG) -cover

.PHONY: mockgen
mockgen: ## generate mock go files
	mockgen -destination=repository/user/mock/mock_user_repo.go -package=mock -source=repository/user/user_repo.go
	mockgen -destination=repository/user/mock/mock_user_redis.go -package=mock -source=repository/user/user_redis.go
	mockgen -destination=controller/user/mock/mock_user.go -package=mock -source=controller/user/user_controller.go
	mockgen -destination=gateway/mock/mock_registry_client.go -package=mock -source=gateway/registry.go

# ==============================================================================
# Modules support

.PHONY: deps-reset
deps-reset:  ## reset dependencies to align with git remote branch
	git checkout -- go.mod
	go mod tidy
	go mod vendor

.PHONY: deps-tidy
deps-tidy:   ## tidy up dependencies and update vendor folder
	go mod tidy
	go mod vendor

.PHONY: deps-upgrade
deps-upgrade:  ## update dependencies to latest version
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -t -d -v ./...
	go mod tidy
	go mod vendor

.PHONY: deps-cleancache
deps-cleancache:  ## remove entire module cache
	go clean -modcache

# ==============================================================================
# Tools commands

.PHONY: lint
lint:  ## run linter
	echo "Starting linters"
	CGO_ENABLED=0 golangci-lint run ./...


# ==============================================================================
# Docker compose commands
.PHONY: docker-protos  # generate protobuf Go code
docker-protos: ## generate code for protobuf
	etc/script/build_protos.sh

.PHONY: docker-local
docker-local:  ## start docker compose for local environment
	echo "Starting local environment"
	docker compose -f docker-compose.yml up --build

