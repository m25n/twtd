package main

import (
	"context"
	"flag"
	"github.com/m25n/twt"
	"github.com/m25n/twt/logger"
	"github.com/m25n/twt/task"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
)

var (
	version   = "0.0.0"
	gitCommit = "unknown"
)

func main() {
	l := log.Default()

	l.Printf("running twtd/%s (%s)", version, gitCommit)
	addr := flag.String("http", ":8080", "address and port to bind to")
	basedir := flag.String("dir", ".", "directory where the twtxt.txt file is located")
	flag.Parse()

	if len(os.Getenv("TWTD_USR")) == 0 || len(os.Getenv("TWTD_PWD")) == 0 {
		l.Fatal("error: You must supply basic auth credentials using the TWTD_USR and TWTD_PWD environment variables")
	}

	l.Printf("storing files in %s", *basedir)
	db, err := twt.NewFileDB(*basedir)
	if err != nil {
		l.Fatalf("error initialize database: %s", err.Error())
	}

	numWorkers := int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	runner := task.NewRunner(numWorkers)
	defer runner.Stop()

	l.Printf("listening on %s", *addr)
	s := &http.Server{
		Addr:    *addr,
		Handler: twt.Handler(logger.New(l), db, twt.BasicAuth(os.Getenv("TWTD_USR"), os.Getenv("TWTD_PWD")), runner.Enqueue),
	}
	defer s.Shutdown(context.Background())
	if err := s.ListenAndServe(); err != nil {
		l.Fatal(err.Error())
	}
}
