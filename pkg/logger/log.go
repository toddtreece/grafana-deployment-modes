package logger

import (
	"context"
	"fmt"

	"github.com/grafana/dskit/services"
)

type Logger interface {
	Log(string)
}

// localLogger is a logger that logs to stdout.
// it is used for target "all".
type localLogger struct {
	*services.BasicService
}

func NewLocalLogger() *localLogger {
	l := &localLogger{}
	l.BasicService = services.NewBasicService(l.start, l.run, l.stop)
	return l
}

func (l *localLogger) start(ctx context.Context) error {
	return nil
}

func (l *localLogger) run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (l *localLogger) stop(failure error) error {
	return failure
}

func (l *localLogger) Log(msg string) {
	fmt.Println(msg)
}
