# デフォルトのポート番号
port?=3333

all: start

# もし競合する場合は「make port=5555」など出来る
start:
	uv run uvicorn main:app --port $(port) --reload