# default
PORT ?= 8000

# make PORT=4000 と指定することもできる
all:
	uv run fastapi dev main.py --port $(PORT)

# 依存関係を消してから入れ直す
re:
	rm -rf .venv
	uv sync

clean:
	rm -rf .venv
