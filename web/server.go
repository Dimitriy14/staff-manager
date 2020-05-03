package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"

	"github.com/Dimitriy14/staff-manager/logger"
	"github.com/urfave/negroni"
)

func NewServer(addr string, handler http.Handler, log logger.Logger, signal chan os.Signal) *server {
	recovery := negroni.NewRecovery()
	negroniLog := negroni.NewLogger()
	negroniLog.ALogger = logger.NewNegroniLogger(log)

	middlewareManger := negroni.New()
	middlewareManger.Use(recovery)
	middlewareManger.Use(negroniLog)
	middlewareManger.UseHandler(handler)

	port := os.Getenv("PORT")
	return &server{
		server: http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: middlewareManger,
		},
		signal: signal,
	}
}

type server struct {
	server http.Server
	signal chan os.Signal
}

func (s *server) Start() {
	go func() {
		err := s.server.ListenAndServe()
		if err != nil {
			log.Printf("Stop listening due to %s", err)
			s.signal <- syscall.SIGQUIT
		}
	}()
}

func (s *server) Stop() error {
	return s.server.Close()
}
