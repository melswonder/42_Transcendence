FROM python:3.12-slim

WORKDIR /app

# uvをインストール
RUN pip install uv

# 依存関係ファイルをコピー
COPY pyproject.toml uv.lock* ./

# 依存関係をインストール
RUN uv sync --no-dev || uv sync

# アプリケーションコードをコピー
COPY . .

# ポートを公開
EXPOSE 8000

# FastAPIサーバーを起動（開発モード）
CMD ["uv", "run", "fastapi", "dev", "main.py", "--host", "0.0.0.0", "--port", "8000"]