package client

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() *zap.Logger {
	return zap.New(&SlogCore{
		Encoder: zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
	})
}

type SlogCore struct {
	Encoder zapcore.Encoder
}

func (c *SlogCore) Enabled(level zapcore.Level) bool {
	var lvl slog.Level
	switch level {
	case zapcore.DebugLevel:
		lvl = slog.LevelDebug
	case zapcore.WarnLevel:
		lvl = slog.LevelWarn
	case zapcore.ErrorLevel, zapcore.FatalLevel, zapcore.PanicLevel:
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	return slog.Default().Enabled(context.Background(), lvl)
}

func (c *SlogCore) With(fields []zapcore.Field) zapcore.Core {
	clone := *c
	clone.Encoder = c.Encoder.Clone()
	return &clone
}

func (c *SlogCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}

	return ce
}

func (c *SlogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	var slogFields []any

	for _, f := range fields {
		var v any
		switch f.Type {
		case zapcore.ByteStringType:
			v = f.Interface.([]byte)
		case zapcore.DurationType:
			v = time.Duration(f.Integer)
		case zapcore.Float64Type:
			v = appendFloat(math.Float64frombits(uint64(f.Integer)), 64)
		case zapcore.Float32Type:
			v = appendFloat(float64(math.Float32frombits(uint32(f.Integer))), 32)
		case zapcore.Int64Type:
			v = int64(f.Integer)
		case zapcore.Int32Type:
			v = int32(f.Integer)
		case zapcore.Int16Type:
			v = int16(f.Integer)
		case zapcore.Int8Type:
			v = int8(f.Integer)
		case zapcore.TimeFullType:
			v = f.Interface.(time.Time)
		case zapcore.Uint64Type:
			v = uint64(f.Integer)
		case zapcore.Uint32Type:
			v = uint32(f.Integer)
		case zapcore.Uint16Type:
			v = uint16(f.Integer)
		case zapcore.Uint8Type:
			v = uint8(f.Integer)
		case zapcore.UintptrType:
			v = uintptr(f.Integer)
		case zapcore.ReflectType:
			v = f.Interface
		case zapcore.StringerType:
			v = encodeStringer(f.Key, f.Interface)
		case zapcore.ErrorType:
			v = f.Interface.(error)
		case zapcore.StringType:
			v = f.String
		case zapcore.BoolType:
			v = f.Integer == 1
		default:
			v = f.Interface
		}

		slogFields = append(slogFields, slog.Any(f.Key, v))
	}

	if entry.LoggerName != "" {
		slogFields = append(slogFields, slog.String("type", entry.LoggerName))
	}

	switch entry.Level {
	case zapcore.DebugLevel:
		slog.Debug(entry.Message, slogFields...)
	case zapcore.WarnLevel:
		slog.Warn(entry.Message, slogFields...)
	case zapcore.ErrorLevel, zapcore.FatalLevel, zapcore.PanicLevel:
		slog.Error(entry.Message, slogFields...)
	default:
		slog.Info(entry.Message, slogFields...)
	}

	return nil
}

func (c *SlogCore) Sync() error {
	return nil
}

func encodeStringer(key string, stringer interface{}) (val string) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(stringer); v.Kind() == reflect.Ptr && v.IsNil() {
				val = "<nil>"
				return
			}

			val = fmt.Sprintf("PANIC=%v", err)
		}
	}()

	val = stringer.(fmt.Stringer).String()
	return
}

func appendFloat(val float64, bitSize int) string {
	switch {
	case math.IsNaN(val):
		return `"NaN"`
	case math.IsInf(val, 1):
		return `"+Inf"`
	case math.IsInf(val, -1):
		return `"-Inf"`
	default:
		var bts []byte
		bts = strconv.AppendFloat(bts, val, 'f', -1, bitSize)
		return string(bts)
	}
}
