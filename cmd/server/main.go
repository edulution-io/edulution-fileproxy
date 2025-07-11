package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/edulution-io/edulution-fileproxy/modules/auth"
	"github.com/edulution-io/edulution-fileproxy/modules/config"
	"github.com/edulution-io/edulution-fileproxy/modules/logging"
	"github.com/edulution-io/edulution-fileproxy/modules/smb"
	"github.com/edulution-io/edulution-fileproxy/modules/webdav"
	"github.com/sirupsen/logrus"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	showVersion := flag.Bool("v", false, "Show version")
	configPath := flag.String("c", "/etc/edulution-fileproxy/config.yml", "Path of the configuration-file")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Version: %s (%s)", version, buildDate)
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	logging.Setup(cfg.Log.Level, cfg.Log.File)

	shares, _ := smb.ListShares(cfg)

	handler := webdav.NewHandler(cfg, shares)
	mux := http.NewServeMux()
	mux.Handle(cfg.HTTP.WebDAVPrefix,
		auth.LDAPMiddleware(cfg,
			smb.SessionMiddleware(cfg, handler),
		),
	)
	logrus.Infof("Starting edulution-fileproxy at 0.0.0.0%s!", cfg.HTTP.Address)
	logrus.Fatal(http.ListenAndServeTLS(cfg.HTTP.Address, cfg.HTTP.CertFile, cfg.HTTP.KeyFile, mux))
}
