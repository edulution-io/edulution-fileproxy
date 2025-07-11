package webdav

import (
	"net/http"

	"github.com/edulution-io/edulution-fileproxy/modules/config"
	"github.com/edulution-io/edulution-fileproxy/modules/smb"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/webdav"
)

func NewHandler(cfg *config.Config, shares []smb.Share) http.Handler {
	logrus.Debugf("Creating handler for shares...")
	share_map := make(map[string]smb.FS, len(shares))
	for _, s := range shares {
		logrus.Debugf("* Add share %s to share_map", s.ShareName)
		share_map[s.ShareName] = smb.FS{Share: s.ShareName}
	}

	fs := &RouterFS{Prefix: cfg.HTTP.WebDAVPrefix, Shares: share_map}

	return &webdav.Handler{
		Prefix:     cfg.HTTP.WebDAVPrefix,
		FileSystem: fs,
		LockSystem: webdav.NewMemLS(),
	}
}
