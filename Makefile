GO_PROJECT_NAME := apbp

go_prep_build:
	@echo "\n.... Preparing installation environment for $(GO_PROJECT_NAME).... "
	go get github.com/cespare/reflex

go_dep_install:
	@echo "\n.... Installing dependencies for $(GO_PROJECT_NAME)...."
	go mod tidy
	go mod download

go_build:
	@echo "\n.... Building $(GO_PROJECT_NAME)...."
	go build -o ./bin/api ./cmd/api

go_migrate:
	@echo "\n... Migrate db schemas and validations $(GO_PROJECT_NAME)...."
	go build -o ./bin/migrate ./cmd/migration && ./bin/migrate

go_run:
	@echo "\n.... Running $(GO_PROJECT_NAME)...."
	./bin/api

# Project rules
install:
	$(MAKE) go_prep_build
	$(MAKE) go_dep_install
	$(MAKE) go_build

init:
ifeq ($(migrate), true)
	$(MAKE) go_migrate
endif
ifeq ($(dev), true)
	reflex -sr '\.go$$' -- make restart
else
	$(MAKE) go_build
	$(MAKE) go_run
endif

restart:
	@$(MAKE) go_dep_install
	@$(MAKE) go_build
	@$(MAKE) go_run

run: 
ifeq ($(dev), true)
	@echo "dev=$(dev)\migrate=$(migrate)" > .env.docker && docker-compose up --force-recreate
endif
	@echo "dev=$(dev)\migrate=$(migrate)" > .env.docker && docker-compose up --force-recreate -d 

clear:
	docker-compose down


.PHONY: go_prep_build go_migrate go_dep_install go_build go_run install run restart reflex