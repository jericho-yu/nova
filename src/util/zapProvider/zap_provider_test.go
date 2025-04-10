package zapProvider

import (
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Test1(t *testing.T) {
	t.Run("test1 创建日志", func(t *testing.T) {
		var (
			zapLogger *zap.Logger
			err       error
		)

		if zapLogger, err = ZapProviderApp.New(
			ZapProviderConfig.New(zapcore.ErrorLevel).
				SetPath(".").
				SetPathAbs(false).
				SetInConsole(false).
				SetEncoderType(EncoderTypeConsole).
				SetNeedCompress(true).
				SetMaxBackup(30).
				SetMaxSize(10).
				SetMaxDay(30),
		); err != nil {
			t.Fatal(err)
		}

		zapLogger.Info("test-info", zap.String("a", "b"))
		zapLogger.Debug("test-debug", zap.String("c", "d"))
		zapLogger.Warn("test-warning", zap.Any("any", []any{"haha", "hehe", 1, 2, 3, 4}))
		zapLogger.Error("test-error", zap.Errors("errors", []error{errors.New("err1"), errors.New("err2"), errors.New("err3")}))
	})
}
