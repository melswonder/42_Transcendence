COMPOSE_FILE = docker-compose.yml

all: build up

build:
	docker compose -f $(COMPOSE_FILE) build

up:
	docker compose -f $(COMPOSE_FILE) up

# ボリューム化していないデータは消える
down:
	docker compose -f $(COMPOSE_FILE) down

stop:
	docker compose -f $(COMPOSE_FILE) stop

start:
	docker compose -f $(COMPOSE_FILE) start

restart: down up

logs:
	docker compose -f $(COMPOSE_FILE) logs -f

# このプロジェクトのコンテナとイメージのみ削除（ボリュームは保持）
clean: down
	docker rmi -f transcendence-db transcendence-backend transcendence-frontend 2>/dev/null || true

# このプロジェクトのコンテナ、イメージ、ボリューム、ネットワークを全て削除
fclean: down
	docker rmi -f transcendence-db transcendence-backend transcendence-frontend 2>/dev/null || true
	docker volume rm postgres_data 2>/dev/null || true
	docker network rm transcendence-network 2>/dev/null || true

re: fclean all

status:
	docker compose -f $(COMPOSE_FILE) ps

images:
	docker images

exec-db:
	docker exec -it postgres psql -U postgres -d transcendence

exec-backend:
	docker exec -it backend sh

exec-frontend:
	docker exec -it frontend bash

# 開発用: frontend をホットリロードの dev サーバーで動かす。
# 既定（make up）は本番ビルド。
COMPOSE_DEV = -f docker-compose.yml -f docker-compose.dev.yml

build-dev:
	docker compose $(COMPOSE_DEV) build

up-dev:
	docker compose $(COMPOSE_DEV) up

down-dev:
	docker compose $(COMPOSE_DEV) down

re-dev: down-dev build-dev up-dev

.PHONY: all build up down stop start restart logs clean fclean re status \
	exec-postgres exec-backend exec-frontend \
	build-dev up-dev down-dev re-dev
