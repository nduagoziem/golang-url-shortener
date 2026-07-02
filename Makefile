# Variables
MIGRATIONS_DIR=./migrations
DB_URL="postgres://nduagoziem:123456789@localhost:5434/postgres?sslmode=disable"
# This finds your Go binary path automatically
GOBIN=$(shell go env GOPATH)/bin
MIGRATE=$(GOBIN)/migrate

.PHONY: migrate-create
# migrate-create
mc:
	@read -p "Enter migration name: " name; \
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

.PHONY: migrate-up
# migrate-up
mp:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

.PHONY: migrate-down
# migrate-down
md:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
.PHONY: migrate-drop
# migrate-drop
mdp:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" drop
export PATH=$PATH:$(go env GOPATH)/bin