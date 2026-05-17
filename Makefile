.DEFAULT_GOAL := help

KEYS_DIR := infra/keys
PRIVATE_KEY := $(KEYS_DIR)/private.pem
PUBLIC_KEY := $(KEYS_DIR)/public.pem
BIN := bin/auth-service

.PHONY: help init keys env build run tidy fmt vet test up down logs clean

help:
	@echo "Targets:"
	@echo "  init    - generate keys + create .env (bootstrap to run)"
	@echo "  keys    - generate RSA keypair into $(KEYS_DIR)"
	@echo "  env     - create .env from .env.example if missing"
	@echo "  build   - build binary into $(BIN)"
	@echo "  run     - run the service locally"
	@echo "  tidy    - go mod tidy"
	@echo "  fmt     - go fmt ./..."
	@echo "  vet     - go vet ./..."
	@echo "  test    - go test ./..."
	@echo "  up      - docker compose up --build -d"
	@echo "  down    - docker compose down"
	@echo "  logs    - follow app logs"
	@echo "  clean   - remove generated artifacts"

init: keys env

keys: $(PUBLIC_KEY)

$(PRIVATE_KEY):
	mkdir -p $(KEYS_DIR)
	openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out $(PRIVATE_KEY)
	chmod 644 $(PRIVATE_KEY)

$(PUBLIC_KEY): $(PRIVATE_KEY)
	openssl rsa -in $(PRIVATE_KEY) -pubout -out $(PUBLIC_KEY)
	chmod 644 $(PUBLIC_KEY)

env:
	@test -f .env || (cp .env.example .env && echo "created .env")

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN) ./cmd/app

run:
	go run ./cmd/app

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f app

clean:
	rm -rf bin $(KEYS_DIR)
