package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func setupLogger() zerolog.Logger {
	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal().Err(err).Msg("无法创建日志目录")
	}

	// 配置日志轮转
	logFile := &lumberjack.Logger{
		Filename:   filepath.Join("logs", "app.log"),
		MaxSize:    100,  // MB
		MaxBackups: 5,    // 保留的旧日志文件数量
		MaxAge:     30,   // 天数
		Compress:   true, // 压缩旧日志
	}

	// 控制台输出配置
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	// 多路输出 - 同时输出到控制台和文件
	multiWriter := zerolog.MultiLevelWriter(consoleWriter, logFile)

	// 构建日志记录器
	logger := zerolog.New(multiWriter).
		With().
		Timestamp().
		Caller().
		Logger()

	// 设置全局日志级别
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	return logger
}
