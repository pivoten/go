package temporal

import (
	tlog "go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

// zapAdapter adapts a *zap.Logger to the Temporal SDK log.Logger interface.
type zapAdapter struct{ z *zap.SugaredLogger }

// NewZapAdapter wraps a zap logger for use as a Temporal SDK logger.
func NewZapAdapter(z *zap.Logger) tlog.Logger {
	if z == nil {
		z = zap.NewNop()
	}
	return &zapAdapter{z: z.Sugar()}
}

func (a *zapAdapter) Debug(msg string, kv ...interface{}) { a.z.Debugw(msg, kv...) }
func (a *zapAdapter) Info(msg string, kv ...interface{})  { a.z.Infow(msg, kv...) }
func (a *zapAdapter) Warn(msg string, kv ...interface{})  { a.z.Warnw(msg, kv...) }
func (a *zapAdapter) Error(msg string, kv ...interface{}) { a.z.Errorw(msg, kv...) }
