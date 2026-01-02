# CMS API Server

Go によるシンプルな CMS API サーバー

## コンセプト

- ローカル専用の記事管理システム
- 記事は Markdown で執筆
- 静的 HTML を生成し、GitHub Pages へ手動デプロイ

## 技術スタック

- Go
- SQLite3
- Gin (Web フレームワーク)

## アーキテクチャ

### レイヤー構成

```
Handler → Service → Repository → DB
```

- **Handler**: HTTP リクエスト/レスポンスの処理
- **Service**: ビジネスロジック
- **Repository**: DB アクセス

### ルール

- Handler は Repository を直接触らない
- Handler は Service のみを呼び出す
- Service は Repository を通じて DB にアクセスする

### ディレクトリ構成

```
internal/
  {domain}/
    model.go       # エンティティ定義
    repository.go  # DBアクセス
    service.go     # ビジネスロジック
    handler.go     # HTTPハンドラ
```

## エンティティ設計

### User（ユーザー）

| カラム        | 型       | 説明                       |
| ------------- | -------- | -------------------------- |
| id            | INTEGER  | PK                         |
| email         | TEXT     | メールアドレス（ユニーク） |
| password_hash | TEXT     | パスワードハッシュ         |
| name          | TEXT     | 表示名                     |
| created_at    | DATETIME | 作成日時                   |
| updated_at    | DATETIME | 更新日時                   |

### Article（記事）

| カラム       | 型       | 説明                      |
| ------------ | -------- | ------------------------- |
| id           | INTEGER  | PK                        |
| title        | TEXT     | タイトル                  |
| slug         | TEXT     | URL スラッグ（ユニーク）  |
| content      | TEXT     | 本文（Markdown）          |
| status       | TEXT     | draft / published         |
| author_id    | INTEGER  | FK → User                 |
| category_id  | INTEGER  | FK → Category（nullable） |
| published_at | DATETIME | 公開日時                  |
| created_at   | DATETIME | 作成日時                  |
| updated_at   | DATETIME | 更新日時                  |

### Category（カテゴリ）

| カラム     | 型       | 説明                     |
| ---------- | -------- | ------------------------ |
| id         | INTEGER  | PK                       |
| name       | TEXT     | カテゴリ名               |
| slug       | TEXT     | URL スラッグ（ユニーク） |
| created_at | DATETIME | 作成日時                 |

### Tag（タグ）

| カラム     | 型       | 説明                     |
| ---------- | -------- | ------------------------ |
| id         | INTEGER  | PK                       |
| name       | TEXT     | タグ名（ユニーク）       |
| slug       | TEXT     | URL スラッグ（ユニーク） |
| created_at | DATETIME | 作成日時                 |

### ArticleTag（記事とタグの中間テーブル）

| カラム     | 型      | 説明         |
| ---------- | ------- | ------------ |
| article_id | INTEGER | FK → Article |
| tag_id     | INTEGER | FK → Tag     |

### Template（テンプレート）

| カラム     | 型       | 説明                       |
| ---------- | -------- | -------------------------- |
| id         | INTEGER  | PK                         |
| name       | TEXT     | テンプレート名（ユニーク） |
| content    | TEXT     | テンプレート内容（HTML）   |
| created_at | DATETIME | 作成日時                   |
| updated_at | DATETIME | 更新日時                   |

## API エンドポイント

### 記事

| Method | Path              | 説明     |
| ------ | ----------------- | -------- |
| GET    | /api/articles     | 記事一覧 |
| GET    | /api/articles/:id | 記事取得 |
| POST   | /api/articles     | 記事作成 |
| PUT    | /api/articles/:id | 記事更新 |
| DELETE | /api/articles/:id | 記事削除 |

### カテゴリ

| Method | Path                | 説明         |
| ------ | ------------------- | ------------ |
| GET    | /api/categories     | カテゴリ一覧 |
| POST   | /api/categories     | カテゴリ作成 |
| PUT    | /api/categories/:id | カテゴリ更新 |
| DELETE | /api/categories/:id | カテゴリ削除 |

### タグ

| Method | Path          | 説明     |
| ------ | ------------- | -------- |
| GET    | /api/tags     | タグ一覧 |
| POST   | /api/tags     | タグ作成 |
| PUT    | /api/tags/:id | タグ更新 |
| DELETE | /api/tags/:id | タグ削除 |

### テンプレート

| Method | Path                        | 説明                             |
| ------ | --------------------------- | -------------------------------- |
| GET    | /api/templates              | テンプレート一覧                 |
| GET    | /api/templates/:name        | テンプレート取得                 |
| PUT    | /api/templates/:name        | テンプレート更新（JSON）         |
| POST   | /api/templates/:name/upload | テンプレートファイルアップロード |
| POST   | /api/templates/import       | ZIP一括インポート                |
| POST   | /api/templates/reset        | 全テンプレートをデフォルトに戻す |

**テンプレート名:**

- `base` - ベーステンプレート（HTML の骨格）
- `article` - 記事個別ページ
- `index` - 記事一覧ページ
- `category` - カテゴリ別一覧ページ
- `tag` - タグ別一覧ページ

### エクスポート（静的サイト生成）

| Method | Path        | 説明                                               |
| ------ | ----------- | -------------------------------------------------- |
| POST   | /api/export | published 記事を HTML 化して出力ディレクトリに生成 |

## 静的サイト生成

### 概要

- `POST /api/export` で静的ファイルを生成
- Markdown → HTML 変換
- 生成後、出力先を GitHub Pages リポジトリとして手動 push

### テンプレート

テンプレートは **データベースで管理** されます。

- 初回起動時にデフォルトテンプレートが自動投入される
- API または管理画面からカスタマイズ可能
- `POST /api/templates/reset` でデフォルトに戻せる

**テンプレートのインポート例（CLI）:**

```bash
# 単一ファイルをアップロード
curl -X POST http://localhost:8080/api/templates/base/upload \
  -F "file=@base.html"

# ZIP一括インポート
curl -X POST http://localhost:8080/api/templates/import \
  -F "file=@templates.zip"

# デフォルトにリセット
curl -X POST http://localhost:8080/api/templates/reset
```

**テンプレートの編集例（CLI）:**

```bash
# テンプレート一覧を取得
curl http://localhost:8080/api/templates

# 特定のテンプレートを取得
curl http://localhost:8080/api/templates/base

# テンプレートを更新
curl -X PUT http://localhost:8080/api/templates/base \
  -H "Content-Type: application/json" \
  -d '{"content": "<!DOCTYPE html>..."}'

# デフォルトにリセット
curl -X POST http://localhost:8080/api/templates/reset
```

**テンプレート変数:**

| テンプレート | 使用可能な変数                                 |
| ------------ | ---------------------------------------------- |
| base         | `{{.Title}}`, `{{.SiteTitle}}`, `{{.Content}}` |
| article      | `{{.Article}}`, `{{.Content}}`                 |
| index        | `{{.Articles}}`                                |
| category     | `{{.Category}}`, `{{.Articles}}`               |
| tag          | `{{.Tag}}`, `{{.Articles}}`                    |

テンプレートは Go の `html/template` 形式。

### 出力先

設定ファイル（`config.json`）で指定：

```json
{
  "export_dir": "/home/kazu/dev/my-blog"
}
```

### 生成されるページ

| ページ         | パス                      | 優先度 |
| -------------- | ------------------------- | ------ |
| 記事個別       | `/posts/{slug}.html`      | Phase1 |
| 記事一覧       | `/index.html`             | Phase1 |
| カテゴリ別一覧 | `/categories/{slug}.html` | Phase2 |
| タグ別一覧     | `/tags/{slug}.html`       | Phase2 |

### 出力ディレクトリ構成

```
{export_dir}/
├── index.html
├── posts/
│   ├── hello-world.html
│   └── second-post.html
├── categories/
│   └── tech.html
└── tags/
    └── go.html
```

## 設定ファイル

`config.json`：

```json
{
  "export_dir": "./dist"
}
```

## 起動方法

```bash
go run main.go
```
