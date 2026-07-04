# ifneq (,$(wildcard ./.env))
    include .env
#     export
# endif


# Variables
MIGRATIONS_DIRECTORY=$(MIGRATIONS_DIR)
DB_URL=$(DATABASE_URL)
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