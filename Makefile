COMPOSE_FILE = docker-compose.yml

all: build up

build:
	docker compose -f $(COMPOSE_FILE) build

# 構築 起動
up:
	docker compose -f $(COMPOSE_FILE) up -d

# 停止 削除　ボリューム化してないものは消える
down:
	docker compose -f $(COMPOSE_FILE) down

# 一時停止
stop:
	docker compose -f $(COMPOSE_FILE) stop

# 再開
start:
	docker compose -f $(COMPOSE_FILE) start

# 再起動
restart: down up

# 各コンテナの
logs:
	docker compose -f $(COMPOSE_FILE) logs -f

# このプロジェクトのコンテナとイメージのみ削除（ボリュームは保持）
clean: down
	docker rmi -f postgres python typescript 2>/dev/null || true

# このプロジェクトのコンテナ、イメージ、ボリューム、ネットワークを全て削除
fclean: down
	docker rmi -f postgres python typescript 2>/dev/null || true
	docker volume rm postgres_data 2>/dev/null || true
	docker network rm transcendence-network 2>/dev/null || true

re: fclean all

# 状態
status:
	docker compose -f $(COMPOSE_FILE) ps

# イメージ
images:
	docker images

# shellに入る
exec-postgres:
	docker exec -it postgres bash

exec-python:
	docker exec -it python bash

exec-typescript:
	docker exec -it typescript bash

.PHONY: all build up down stop start restart logs clean fclean re status exec-postgres exec-python exec-typescript