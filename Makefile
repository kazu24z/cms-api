.PHONY: build run serve import export migrate migrate-down migrate-create schema-dump clean help

# 変数（環境変数で上書き可能）
DB_DRIVER ?= sqlite3
DB_FILE ?= cms.db
DB_URL ?= $(DB_DRIVER)://$(DB_FILE)
MIGRATIONS_DIR := db/migrations/$(DB_DRIVER)
BINARY := cms

# PostgreSQL の場合は .env や export で設定:
#   export DB_DRIVER=postgres
#   export DB_URL="postgres://user:pass@host:5432/dbname"

# ビルド
build:
	go build -o $(BINARY) .

# サーバー実行（旧run互換）
run: build
	./$(BINARY) serve

# サーバー実行
serve: build
	./$(BINARY) serve

# Markdownインポート
# 使用例: make import FILE=article.md
import: build
	./$(BINARY) import $(FILE)

# HTMLエクスポート
# 使用例: make export OUTPUT=./dist
export: build
	./$(BINARY) export -o $(or $(OUTPUT),./dist)

# マイグレーション実行（CLI）※ golang-migrate CLI が必要
migrate-cli:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

# ロールバック（1つ戻す）
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
	@$(MAKE) schema-dump

# 新規マイグレーション作成
migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

# スキーマダンプ
schema-dump:
ifeq ($(DB_DRIVER),sqlite3)
	sqlite3 $(DB_FILE) ".schema" > db/schema.sql
else ifeq ($(DB_DRIVER),postgres)
	pg_dump --schema-only "$(DB_URL)" > db/schema.sql
endif
	@echo "Schema dumped to db/schema.sql"

# DBリセット（SQLite専用：削除して再作成）
db-reset:
	@read -p "Delete $(DB_FILE) and recreate? [y/N]: " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		rm -f $(DB_FILE); \
		$(MAKE) run; \
	else \
		echo "Cancelled."; \
	fi

# クリーン
clean:
	rm -f $(BINARY)

# ヘルプ
help:
	@echo "Usage:"
	@echo "  make build          - Build the binary"
	@echo "  make run            - Run the server (alias: serve)"
	@echo "  make serve          - Run the server"
	@echo "  make import FILE=x  - Import markdown file to DB"
	@echo "  make export         - Export articles to HTML (OUTPUT=./dist)"
	@echo "  make migrate-cli    - Run migrations (via CLI, requires golang-migrate)"
	@echo "  make migrate-down   - Rollback one migration"
	@echo "  make migrate-create - Create new migration file"
	@echo "  make schema-dump    - Dump schema to db/schema.sql"
	@echo "  make db-reset       - Delete DB and recreate"
	@echo "  make clean          - Remove binary"
	@echo ""
	@echo "CLI Commands (after build):"
	@echo "  ./cms serve         - Start API server"
	@echo "  ./cms import <file> - Import markdown to DB"
	@echo "  ./cms export        - Export to HTML"
