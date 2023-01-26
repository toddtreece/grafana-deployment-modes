package sender

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/grafana/dskit/services"
	"github.com/grafana/grafana-deployment-modes/pkg/logger"
)

var _ services.Service = (*Sender)(nil)

var senderID = uuid.New().String()

// Sender is a service that sends messages to a logger every second.
type Sender struct {
	*services.BasicService
	logger logger.Logger
}

func New(logger logger.Logger) *Sender {
	s := &Sender{logger: logger}
	s.BasicService = services.NewBasicService(s.start, s.run, s.stop)
	return s
}

func (s *Sender) start(ctx context.Context) error {
	s.logger.Log("Starting sender: " + senderID)
	return nil
}

func (s *Sender) run(ctx context.Context) error {
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			s.logger.Log(fmt.Sprintf("Sender %s current time %s", senderID, time.Now().String()))
		}
	}
}

func (s *Sender) stop(failure error) error {
	if failure != nil {
		return failure
	}
	s.logger.Log("Stopping sender: " + senderID)
	return nil
}
