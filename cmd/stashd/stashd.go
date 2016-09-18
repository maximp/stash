package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/maximp/stash/db"
	"github.com/maximp/stash/server"
)

// main implements entry point of stashd command-line application
func main() {
	log := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

	d, err := db.New(db.Config{
		Log:         log,
		QueueLength: 10,
	})
	if err != nil {
		panic(err)
	}
	defer d.Close()

	stop := make(chan struct{})
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig

		signal.Stop(sig)
		close(sig)

		stop <- struct{}{}
		close(stop)
	}()

	handler := func(cmd []byte, arg []byte) ([]byte, error) {
		c, err := db.ParseCommand(cmd)
		if err != nil {
			return nil, err
		}
		return d.Exec(c, arg)
	}

	cfg := server.Config{
		Logger: log,
		Stop:   stop,
	}

	if err := server.ListenAndServe(":7777", handler, &cfg); err != nil {
		log.Println(err)
	} else {
		log.Println("finished")
	}
}
