package smb

import (
	"context"
	"net"
	"net/http"

	"github.com/edulution-io/edulution-fileproxy/modules/auth"
	"github.com/edulution-io/edulution-fileproxy/modules/config"
	"github.com/edulution-io/edulution-fileproxy/modules/types"
	"github.com/hirochachacha/go-smb2"
	"github.com/sirupsen/logrus"
)

func SessionMiddleware(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := auth.ClientIP(r)
		logrus.Debugf("[%s] Establishing SMB session", clientIP)
		val := r.Context().Value("smbCreds")
		creds, ok := val.(*types.SmbCreds)
		if !ok {
			logrus.Errorf("[%s] Missing SMB credentials in context", clientIP)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		connTCP, err := net.Dial("tcp", cfg.SMB.Server)
		if err != nil {
			logrus.Errorf("[%s] Failed to dial SMB server %s: %v", clientIP, cfg.SMB.Server, err)
			http.Error(w, "File server unreachable", http.StatusServiceUnavailable)
			return
		}
		defer connTCP.Close()

		dialer := &smb2.Dialer{Initiator: &smb2.NTLMInitiator{User: creds.User, Password: creds.Pass, Domain: creds.Domain}}
		sess, err := dialer.Dial(connTCP)
		if err != nil {
			logrus.Errorf("[%s] SMB session error for user %s: %v", clientIP, creds.User, err)
			http.Error(w, "File server auth error", http.StatusUnauthorized)
			return
		}
		defer sess.Logoff()
		logrus.Infof("User %s established SMB session from %s", creds.User, clientIP)

		ctx := context.WithValue(r.Context(), "smbSess", sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
