package gorogger_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goroumaru/gorogger"
	"github.com/pelletier/go-toml"
	pkgerrors "github.com/pkg/errors"

	"go.uber.org/zap/zapcore"
)

type ConfigLogger struct {
	LogLevel     string `toml:"log_level"`
	ConsoleLevel string `toml:"console_level"`
	Path         string `toml:"log_path"`
}

type ConfigData struct {
	Logger ConfigLogger `toml:"logger"`
}

func loadConfig() (ConfigData, error) {
	d, err := ioutil.ReadFile("config.toml")
	if err != nil {
		return ConfigData{}, fmt.Errorf("%w", err)
	}
	var cfg ConfigData
	err = toml.Unmarshal(d, &cfg)
	if err != nil {
		return ConfigData{}, fmt.Errorf("%w", err)
	}
	return cfg, nil
}

func deleteFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

func TestLevel(t *testing.T) {
	cases := map[string]struct {
		logLvl     string
		consoleLvl string
	}{
		"LOG,CONS = INFO,ERR":          {"info", "err"},
		"LOG,CONS = ERR,INFO":          {"err", "info"},
		"LOG,CONS = NOT_USED,ERR":      {"", "err"},
		"LOG,CONS = ERR,NOT_USED":      {"err", ""},
		"LOG,CONS = NOT_USED,DEBUG":    {"", "dbg"},
		"LOG,CONS = NOT_USED,WARN":     {"", "warn"},
		"LOG,CONS = NOT_USED,NOT_USED": {"", ""},
	}

	for name, c := range cases {
		path := filepath.Join("testData", name+".json")
		if err := deleteFile(path); err != nil {
			panic(err)
		}

		logger := gorogger.NewLogger(
			path,
			gorogger.GetLevel(c.logLvl),
			gorogger.GetLevel(c.consoleLvl),
		)
		fmt.Printf("---------- %s ----------\n", name)
		logger.Debug("Debug Message", name, "test name")
		logger.Info("Info Message", name, "test name")
		logger.Warn("Warn Message", name, "test name")
		logger.Error("Error Message", name, "test name")
		logger.Close()
	}
}

func TestLogger(t *testing.T) {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}
	if err := deleteFile(cfg.Logger.Path); err != nil {
		panic(err)
	}

	logger := gorogger.NewLogger(
		cfg.Logger.Path,
		gorogger.GetLevel(cfg.Logger.LogLevel),
		gorogger.GetLevel(cfg.Logger.ConsoleLevel),
	)
	defer logger.Close()

	// msgキーへ登録
	logger.Info("No.1: Please write messages here!", nil)

	// "key":"value"形式と表示される
	logger.Info("No.2: Please write messages here!", "value", "key")

	// "key"は省略できる
	logger.Info("No.3: Please write messages here!", "value")

	// "key"は複数してできない。1つ目が出力される
	logger.Info("No.4: Please write messages here!", "value", "key1", "key2")

	// 無名構造体を利用できる
	logger.Info("No.5: Please write messages here!", struct {
		ID   int
		Name string
		age  int // プライペートメンバーは出力されない
	}{
		ID:   1,
		Name: "goroumaru",
		age:  10,
	}, "key")

	// user structを利用できる
	user := &user{
		ID:   1,
		Name: "goroumaru",
		age:  10,
	}
	logger.Error("No.6: Please write messages here!", user, "user struct")

	// slice in structを利用できる
	union := &union{
		ID:      1,
		Members: []string{"goroumaru41gou", "goroumaru42gou"},
	}
	logger.Error("No.6: Please write messages here!", union, "slice in struct")

	// ログ用メソッドを追加すれば、メソッドが実行された結果となる
	adjust := &adjust{
		ID: 1,
		// Timestampはadjustメソッドで入力される
	}
	logger.Error("No.7: Please write messages here!", adjust, "struct with method")

	// 標準errorsによるラッピング
	// NOTE: errorsはスタックトレースに非対応
	logger.Info("No.8: Please write messages here!", doEverything1(), "errors")

	// pkg/errorsによるスタックトレース
	logger.Info("No.9: Please write messages here!", doEverything2(), "pkg/errors")
}

// 標準errorsによるラッピング
func doEverything1() error {
	err := doAnything1()
	return fmt.Errorf("3.Wrapped: %w", err)
}
func doAnything1() error {
	err := doSomething1()
	return fmt.Errorf("2.Wrapped: %w", err)
}
func doSomething1() error {
	return errors.New("1.Occured")
}

// pkg/errorsでエラースタックする
func doEverything2() error {
	err := doAnything2()
	return pkgerrors.WithStack(
		fmt.Errorf("3.Wrapped: %w", err),
	)
}
func doAnything2() error {
	err := doSomething2()
	return pkgerrors.WithStack(
		fmt.Errorf("2.Wrapped: %w", err),
	)
}
func doSomething2() error {
	return pkgerrors.New("1.Occured")
}

type user struct {
	ID   int
	Name string
	age  int // プライペートメンバーは出力されない
}

type union struct {
	ID      int
	Members []string
}

type adjust struct {
	ID        int
	Timestamp int64
}

func (a *adjust) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	// enc.AddInt("ID", a.ID) // Addしないと、デフォルト値になってしまう
	enc.AddInt64("Timestamp", time.Now().UnixMilli())
	return nil
}
