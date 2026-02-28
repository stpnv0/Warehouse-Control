package service

import (
	"context"
	"time"

	"github.com/wb-go/wbf/logger"
)

type nopLogger struct{}

func newTestLogger() logger.Logger            { return &nopLogger{} }
func (n *nopLogger) Debug(string, ...any)     {}
func (n *nopLogger) Info(string, ...any)      {}
func (n *nopLogger) Warn(string, ...any)      {}
func (n *nopLogger) Error(string, ...any)     {}
func (n *nopLogger) Debugw(string, ...any)    {}
func (n *nopLogger) Infow(string, ...any)     {}
func (n *nopLogger) Warnw(string, ...any)     {}
func (n *nopLogger) Errorw(string, ...any)    {}
func (n *nopLogger) Ctx(context.Context) logger.Logger { return n }
func (n *nopLogger) With(...any) logger.Logger         { return n }
func (n *nopLogger) WithGroup(string) logger.Logger    { return n }
func (n *nopLogger) LogRequest(context.Context, string, string, int, time.Duration) {}
func (n *nopLogger) Log(logger.Level, string, ...logger.Attr)                       {}
func (n *nopLogger) LogAttrs(context.Context, logger.Level, string, ...logger.Attr) {}
