package webdav

import (
	"net/http"

	"github.com/edulution-io/edulution-fileproxy/modules/config"
	"github.com/edulution-io/edulution-fileproxy/modules/smb"
	"golang.org/x/net/webdav"
)

func NewHandler(cfg *config.Config) http.Handler {
	return &webdav.Handler{
		Prefix:     cfg.HTTP.WebDAVPrefix,
		FileSystem: smb.FS{Share: cfg.SMB.ShareName},
		LockSystem: webdav.NewMemLS(),
	}
}
