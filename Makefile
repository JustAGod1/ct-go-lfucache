LOCAL_BIN := $(CURDIR)/bin

GOLANGCI_BIN := $(LOCAL_BIN)/golangci-lint
GOLANGCI_TAG=1.61.0

GO_TEST=$(LOCAL_BIN)/gotest
GO_TEST_ARGS="-race -v ./..."

# Выполнить полный цикл
all: lint test

# Устанавливает зависимости для использования
.PHONY: install-deps
install-deps:
	echo 'Installing dependencies...'
	tmp=$$(mktemp -d) && cd $$tmp && pwd && go mod init temp && \
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_TAG) && \
	GOBIN=$(LOCAL_BIN) go install github.com/rakyll/gotest@v0.0.6 && \
	rm -fr $$tmp

# Запускает линтер по файлам
.PHONY: lint
lint: install-deps
	echo 'Running linter on files...'
	$(GOLANGCI_BIN) run \
	--new-from-rev=origin/main \
	--config=.golangci.yaml \
	--sort-results \
	--max-issues-per-linter=0 \
	--max-same-issues=0

# Прогоняет все тесты
.PHONY: test
test: install-deps
	echo 'Running tests...'
	${GO_TEST} "${GO_TEST_ARGS}"

# Обновить репозиторий
.PHONY: update
update:
ifeq (-n $(git status --untracked-files=no --porcelain),)
	@echo 'You have some changes. Please commit, checkout or stash them.'
	@exit 1
endif
	@echo 'Updating repository to latest main...'
	# Обновляем main
	git checkout main
	git pull --all
	# Обновляем hw
	git checkout hw
	git pull --all
	git rebase main
	git push -f
	@echo 'Successfully updated'
