# nvmをダウンロードしてインストールする：
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash

# シェルを再起動する代わりに実行する
\. "$HOME/.nvm/nvm.sh"

# Node.jsをダウンロードしてインストールする：
nvm install 20

# pnpmをダウンロードしてインストールする：
corepack enable pnpm

# 依存関係ダウンロード
cd ../../typescript | pnpm i