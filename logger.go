package gorogger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	mux     sync.Mutex
	log     *zap.Logger
	logFile *os.File
}

func NewLogger(
	logPath string,
	logLevel OutputLevel, // ログファイル出力レベル
	consoleLevel OutputLevel, // コンソール出力レベル
) *Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// ログファイル
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(logPath), os.ModePerm); err != nil {
			panic(err)
		}
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}

	cores := []zapcore.Core{}
	// ログ・コア生成
	if logLevel != NOT_USED {
		logSync := zapcore.AddSync(f)
		cores = append(cores, zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig), // 構造化ログ（JSON）
			zapcore.Lock(logSync),
			setLevel(logLevel),
		))
	}
	// コンソール・コア生成
	if consoleLevel != NOT_USED {
		ioSync := zapcore.AddSync(os.Stdout) // io.Writerをzapcore.WriteSyncerに変換
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig), // コンソール出力
			zapcore.Lock(ioSync),
			setLevel(consoleLevel),
		))
	}

	// コア結合
	log := zap.New(zapcore.NewTee(cores...))
	return &Logger{
		sync.Mutex{},
		log,
		f,
	} // 呼び出し元でdefer Close()メソッドし、ファイルを閉じる
}

func (l *Logger) Close() {
	l.Info("Close logger.", nil)

	l.mux.Lock()
	defer l.mux.Unlock()
	if err := l.syncFile(); err != nil {
		l.Info("Log file can not be synchronized.", zap.Error(err))
	}
	if err := l.closeFile(); err != nil {
		panic(
			fmt.Errorf("log file can not be closed: %w", err),
		)
	}
}

// Debug : use zap logger. If val don't exist, set "nil".
func (l *Logger) Debug(message string, val interface{}, keys ...string) {
	assertFields(l.log.Debug, message, val, keys...)
}

// Info : use zap logger. If val don't exist, set "nil".
func (l *Logger) Info(message string, val interface{}, keys ...string) {
	assertFields(l.log.Info, message, val, keys...)
}

// Warn : use zap logger. If val don't exist, set "nil".
func (l *Logger) Warn(message string, val interface{}, keys ...string) {
	assertFields(l.log.Warn, message, val, keys...)
}

// Error : use zap logger. If val don't exist, set "nil".
func (l *Logger) Error(message string, val interface{}, keys ...string) {
	assertFields(l.log.Error, message, val, keys...)
}

// Flushing the file system's in-memory copy of recently written data to disk.
func (l *Logger) syncFile() error {
	return l.logFile.Sync()
}

// Close log file
func (l *Logger) closeFile() error {
	if l.logFile == nil {
		return nil
	}
	err := l.logFile.Close()
	l.logFile = nil // closeしたので、nilとする
	return err
}

// https://qiita.com/tsurumiii/items/0294feebc0216b185765
func assertFields(log func(msg string, fields ...zapcore.Field), message string, val interface{}, keys ...string) {
	var key string
	if len(keys) != 0 {
		key = keys[0]
	}
	log(message, zap.Any(key, val))
}

func setLevel(logLevel OutputLevel) (level zapcore.Level) {
	switch logLevel {
	case DBG:
		level = zapcore.DebugLevel
	case INFO:
		level = zapcore.InfoLevel
	case WARN:
		level = zapcore.WarnLevel
	case ERR:
		level = zapcore.ErrorLevel
	}
	return
}
