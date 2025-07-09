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
		ip := ClientIP(r)
		logrus.Debugf("[%s] Got new request: %s %s", ip, r.Method, r.URL.RequestURI())
		logrus.Debugf("[%s] Starting LDAP authentication", ip)
		user, pass, ok := r.BasicAuth()
		if !ok {
			logrus.Debugf("[%s] Missing BasicAuth credentials", ip)
			w.Header().Set("WWW-Authenticate", `Basic realm="`+cfg.LDAP.Domain+`"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		l, err := ldap.DialURL(cfg.LDAP.Server, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: cfg.LDAP.InsecureSkipVerify}))
		if err != nil {
			logrus.Errorf("LDAP dial error: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		defer l.Close()
		bindDN := user + "@" + cfg.LDAP.Domain
		if err := l.Bind(bindDN, pass); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		logrus.Infof("User %s authenticated", user)
		creds := types.NewSmbCreds(user, pass, cfg.LDAP.Domain)
		ctx := context.WithValue(r.Context(), "smbCreds", creds)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
