package client

import (
	"fmt"
	"testing"

	"go.uber.org/zap"
)

func Test_Logger(t *testing.T) {
	logger := NewLogger()
	logger.Info("This is a test log.")
	logger.Named("etcd-client").Info("This is a test log.")
	logger.Info("This is a test log with fields.", zap.Int("test", 1))
	logger.Info("This is a test log with fields.", zap.Float32("test", 1.12))
	logger.Info("This is a test log with fields.", zap.String("test", "test"))
	logger.Info("This is a test log with fields.", zap.Strings("test", []string{"test", "test1"}))
	logger.Info("This is a test log with fields.", zap.Error(fmt.Errorf("test")))
}
