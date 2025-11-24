package main

import (
	"context"
	"fmt"
	"os"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/config"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/ports"
	log "github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
)

type options struct {
	Port int    `short:"p" long:"port" description:"port to listen" default:"8080"`
	Conf string `short:"f" long:"conf" env:"FM_CONF" default:"for_twenty_readers.yml" description:"config file (yml)"`
	Dbg  bool   `long:"dbg" env:"DEBUG" description:"debug mode"`
}

var revision = "local"

func main() {
	fmt.Printf("For twenty readers %s\n", revision)
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}
	setupLog(opts.Dbg)

	ctx := context.Background()

	srv := &ports.Server{
		Version: revision,
		Conf:    config.Conf{},
	}
	srv.Run(ctx, opts.Port)
}

func setupLog(dbg bool) {
	if dbg {
		log.Setup(log.Debug, log.CallerFile, log.Msec, log.LevelBraces)
		return
	}
	log.Setup(log.Msec, log.LevelBraces)
}
