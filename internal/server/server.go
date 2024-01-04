package server

import (
	"context"

	"github.com/labstack/echo/v4"
)

type Server struct {
	E    *echo.Echo
	addr string
}

func NewServer(addr string, router *echo.Echo) *Server {
	return &Server{E: router, addr: addr}
}

func (s *Server) Run() error {
	return s.E.Start(s.addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.E.Shutdown(ctx)
}
