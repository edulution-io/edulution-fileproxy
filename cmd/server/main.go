package main

import (
	"flag"
	"net/http"

	"github.com/edulution-io/edulution-fileproxy/modules/auth"
	"github.com/edulution-io/edulution-fileproxy/modules/config"
	"github.com/edulution-io/edulution-fileproxy/modules/logging"
	"github.com/edulution-io/edulution-fileproxy/modules/smb"
	"github.com/edulution-io/edulution-fileproxy/modules/webdav"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "config.yml", "config file path")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		logrus.Fatal(err)
	}
	logging.Setup(cfg.Log.Level, cfg.Log.File)
	handler := webdav.NewHandler(cfg)
	mux := http.NewServeMux()
	mux.Handle(cfg.HTTP.WebDAVPrefix,
		auth.LDAPMiddleware(cfg,
			smb.SessionMiddleware(cfg, handler),
		),
	)
	logrus.Infof("Starting server at %s", cfg.HTTP.Address)
	logrus.Fatal(http.ListenAndServeTLS(cfg.HTTP.Address, cfg.HTTP.CertFile, cfg.HTTP.KeyFile, mux))
}
