Python 3.12.12v

#　基本的な使い方

uv init: 新規プロジェクトを作成 最初しか使わない
uv add: 依存関係を追加 ライブラリを入れる的な感じ
uv remove: 依存関係を削除
uv sync: 仮想環境を更新 ←これだけでいい
uv run: スクリプトを実行
uv lock: ロックファイルを更新
uv pip list: 依存関係のリスト表示

# 仮想環境

```bash
source ./.venv/bin/activate
```

```bash
source ./.venv/bin/activate.fish 
```

# 下やらなくていいかも1/14 uvで自動的にvenvを見ることができるらしく、sourceをわざわざしなくてもいいらしい
# .venvを常に適応するためには
```bash
pwd | xargs -I {} echo "{}/.venv/bin/activate"
```
これの出力結果を
command + shift + P　で
Python: Select Interpreter　を選択
インタプリターパスの選択をする

# 備考
Python 本体 の一覧	uv python list	PC内にどのバージョンの Python があるか確認する
パッケージ の一覧	uv pip list	現在の環境に何のライブラリ（pandasなど）があるか確認する
ツール の一覧	uv tool list	uv でインストールした独立したツール（ruffなど）を確認する