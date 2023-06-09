```
_____  _____  _____  _____  _____  _____  ___  _____  _____  __ __  _____  _____ 
/  _  \/  _  \/   __\/  _  \/  _  \/  _  \/___\<___  \/  _  \/  |  \/  _  \/  _  \
|  |  ||   __/|   __||  |  ||  _  ||   __/|   | /  __/|  _  <|  |  ||  |  ||  |  |
\_____/\__/   \_____/\__|__/\__|__/\__/   \___/<_____|\__|\_/\_____/\__|__/\__|__/
```

## 概要 / General

このユーティリティは、openapiの定義からAPIテストツールである[runn](https://github.com/k1LoW/runn)のテストシナリオを生成するものです。

## インストール / Install

シングルバイナリで稼働するため、インストールは不要です。  
[Releases](https://github.com/calloc134/openapi2runn)からダウンロードしてください。

## 使い方 / Usage

```bash
$ openapi2runn gen -i (OpenAPI定義) -o (シナリオ保存ディレクトリ) -s (サーバのホスト)
```

このコマンドを実行すると、`-o`で指定したディレクトリにシナリオが生成されます。
なお、OpenAPI定義はjsonとyamlの両方が指定可能です。

runnの実行は以下のように行います。なお、カレントディレクトリは`-o`で指定したディレクトリにしてください。

```bash
$ runn run **/**/*.yml (対象サーバのホスト)
```

## 特徴 / Features

### パスとメソッド毎に分けられたディレクトリ配置

生成されるディレクトリ配置は以下のようになります。  
なお、以下は一例です。

```
.
├── /0_base/
│   ├── /ApiProfileMe/
|   |   ├── /GET/
│   │   │   └── base.yml
│   │   ├── /PUT/
│   │   │   └── base.yml
│   │   └── /DELETE/
│   │       └── base.yml
│   └── /ApiPostId/
|       ├── /GET/
│       │   └── base.yml
│       ├── /PUT/
│       │   └── base.yml
│       └── /DELETE/
│           └── base.yml
└── /1_noAuth/
    ├── /ApiProfileMe/
    |   ├── /GET/
    │   │   ├── base.yml
    |   |   └── data.json
    │   ├── /PUT/
    │   │   ├── base.yml
    |   |   └── data.json
    │   └── /DELETE/
    │       ├── base.yml
    |       └── data.json
    └── /ApiPostId/
        ├── /GET/
        │   ├── base.yml
        |   └── data.json
        ├── /PUT/
        │   ├── base.yml
        |   └── data.json
        └── /DELETE/
            ├── base.yml
            ├── data.json
            └── config.toml
``` 
URL内のパラメータも、ディレクトリ名として扱われます。

0_baseディレクトリには、APIのベースとなるテストシナリオが格納されます。

以下に実例を示します。
    
```yaml
if: included
runners:
  req: http://localhost
steps:
  first:
    req:
      /api/auth/login:
        POST:
          body: 
            application/json: "{{ vars.req.body }}"
    test:
      compare(steps.first.res.status, vars.res.status)
```

1_noAuthディレクトリには、認証が不要なAPIのテストシナリオが格納されます。

以下に実例を示します。
    
```yaml
desc:
  (POST) /api/auth/loginのテスト
vars:
  data: "json://data.json"

steps:
  first:
    loop: 
      count: len(vars.data)
    include:
      path:
        ../../../0_base/ApiAuthLogin/POST/base.yml
      vars:
        req: '{{ vars.data[i].req }}'
        res: '{{ vars.data[i].res }}'
```
また、このディレクトリには、テストデータを格納するdata.jsonというファイルがあります。

```json
[
    {
        "req": {
            "body": {"password":"dummy","screenName":"dummy"},
            "query": {}
        },
        "res": {
            "status": 200
        }
    }
]
```
このjsonの配列にデータを追加することで、複数のデータをループしてテストすることが可能です。

## 構成ファイル / Configuration

1_noAuthディレクトリには、構成ファイルとしてconfig.tomlを配置することができます。
この構成ファイルは、既に1_noAuth以下にファイル群が存在する場合に有効です。
このtomlの書き方は以下の通りです。
  
```toml
allowOverride = true
```

この構成ファイルには、以下の項目を設定することができます。
| 項目名 | 説明 | 
----|----
| allowOverride | この項目がtrueの場合、当ディレクトリのdata.jsonを上書きすることができます。 |


## 注意事項 / Caution
- このツールは、現在開発中のため、バグが含まれている可能性があります。

## 今後の予定 / Future Plans

 - [ ] Github Actionsでの自動ビルドとデプロイ
 - [ ] ログイン認証に対応したテストシナリオの生成
 - [ ] テストデータの自動生成(?)











