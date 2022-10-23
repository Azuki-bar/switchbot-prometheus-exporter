# switchbot-prometheus-exporter

## English

### about this

This is prometheus exporter which collects temparetures and humidity from [SwitchBot Thermometer and Hygrometeror](https://www.switchbot.jp/products/switchbot-meter).

_English docs will be opened soon._

### Requirements

### How To Use

### Thanks

## Japanese - 日本語

### about

これは SwitchBot 温湿度計からデータを集める Prometheus Exporter です。

### Requirements

この exporter を用いて温湿度情報を Prometheus に収集させるためには最低限以下の用意が必要です。

- [SwitchBot ハブミニ](https://www.switchbot.jp/collections/start-your-smart-life/products/switchbot-hub-mini)
- [SwitchBot 温湿度計](https://www.switchbot.jp/products/switchbot-meter)
- [SwitchBot OPEN API token](https://github.com/OpenWonderLabs/SwitchBotAPI#getting-started)

### 使い方

デフォルトでは 8080 番の `/metrics`にてメトリクスを公開します。

いくつかの環境変数を設定することが必須です。また特定の環境変数を設定することにより振る舞いを変更することができます。

#### `TOKEN`

[required]

SwitchBot の OPEN API TOKEN を渡します。

取得方法は[公式 API ドキュメントの Getting started](https://github.com/OpenWonderLabs/SwitchBotAPI#getting-started)を参考にしてください。

#### `DEVICE_ID`

[required]

監視対象のデバイス ID です。

非公式ですがデバイスの BLE MAC アドレスを英大文字と数字で構成される 12 桁で表現したものとなっています。
例えば BLE mac アドレスが`00:00:5e:00:53:00`であれば`00005E005300`となります。

#### `FETCH_INTERVAL`

[optional, default=5m]

SwitchBot API から情報を取得する間隔を指定します。フォーマットは[このメソッドの仕様](https://pkg.go.dev/time#ParseDuration)に従います。

#### `PORT`

[optional, default=8080]

メトリクスを公開するポートを指定します。
