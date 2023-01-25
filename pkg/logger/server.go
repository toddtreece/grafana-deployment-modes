package logger

import (
	"context"
	"net"

	"github.com/grafana/dskit/services"
)

const ADDRESS = "localhost:19999"

var _ services.Service = (*Server)(nil)

// Server is a TCP service that listens for log messages
type Server struct {
	*services.BasicService
	logger   *localLogger
	listener net.Listener
}

func NewServer() *Server {
	s := &Server{
		logger: NewLocalLogger(),
	}
	s.BasicService = services.NewBasicService(s.start, s.run, s.stop)
	return s
}

func (s *Server) start(ctx context.Context) error {
	listener, err := net.Listen("tcp", ADDRESS)
	if err != nil {
		return err
	}

	s.listener = listener
	return nil
}

func (s *Server) run(ctx context.Context) error {
	conn, err := s.listener.Accept()
	listenerErr := make(chan error, 1)
	go func() {
		for {
			if err != nil {
				listenerErr <- err
			}
			err = s.handler(conn)
			if err != nil {
				listenerErr <- err
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-listenerErr:
		return err
	}
}

func (s *Server) handler(conn net.Conn) error {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		return err
	}
	s.logger.Log(string(buf))
	return nil
}

func (s *Server) stop(failure error) error {
	defer s.listener.Close()
	if failure != nil {
		return failure
	}
	return nil
}
