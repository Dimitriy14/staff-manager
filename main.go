package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dimitriy14/staff-manager/app"
)

func main() {
	config := flag.String("config", "config.json", "-config path/to/config/file.json")
	flag.Parse()

	s := make(chan os.Signal, 1)
	signal.Notify(s,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	application, err := app.LoadApplication(*config, s)
	if err != nil {
		log.Println(err)
		s <- syscall.SIGQUIT
	}

	<-s

	log.Println("Stopping application")
	application.Stop()
	log.Println("Application has been stopped")
}
