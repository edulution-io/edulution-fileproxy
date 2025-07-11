package auth

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/edulution-io/edulution-fileproxy/modules/config"
	"github.com/edulution-io/edulution-fileproxy/modules/types"
	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
)

func LDAPMiddleware(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Debugf(">>> %s %s", r.Method, r.URL.Path)

		ip := ClientIP(r)
		logrus.Debugf("[%s] Starting LDAP authentication", ip)
		user, pass, ok := r.BasicAuth()
		if !ok {
			logrus.Debugf("[%s] Missing BasicAuth credentials", ip)
			w.Header().Set("WWW-Authenticate", `Basic realm="`+cfg.SMB.Domain+`"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		l, err := ldap.DialURL(cfg.LDAP.Server, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: cfg.LDAP.InsecureSkipVerify}))
		if err != nil {
			logrus.Errorf("[%s] LDAP dial error: %v", ip, err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		defer l.Close()
		bindDN := user + "@" + cfg.SMB.Domain
		if err := l.Bind(bindDN, pass); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		logrus.Infof("[%s] User %s authenticated", ip, user)
		creds := types.NewSmbCreds(user, pass, cfg.SMB.Domain)
		ctx := context.WithValue(r.Context(), "smbCreds", creds)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
