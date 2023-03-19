# gorogger

zap logger を利用してコンソールおよびファイルへログ出力する。  
それぞれに、以下設定可能。

- ログレベル選択  
  `dbg`,`info`,`warn`,`err`, `""`
- Output の Enable/Disable 選択  
  ログレベル選択において`""`選択

## How to run

`logger_test.go`

## Setup

`config.toml`

```
   // config.toml
   // LEVEL = dbg,info,warn,err, ""
   [logger]
   CONSOLE_LEVEL = "dbg"
   LOG_LEVEL = "err"
   LOG_PATH = "./testData/test_log.json"
```

## License

[MIT](https://github.com/goroumaru/gorogger/blob/main/LICENSE)
