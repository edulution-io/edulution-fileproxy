package smb

import (
	"net"

	"github.com/edulution-io/edulution-fileproxy/modules/config"
	"github.com/hirochachacha/go-smb2"
	"github.com/sirupsen/logrus"
)

type Share struct{ ShareName string }

func ListShares(cfg *config.Config) ([]Share, error) {
	logrus.Debugf("Loading shares from fileserver %s", cfg.SMB.Server)

	conn, err := net.Dial("tcp", cfg.SMB.Server)
	if err != nil {
		logrus.Fatalf("Unable to connect to SMB server: %v", err)
		return nil, err
	}
	defer conn.Close()

	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     cfg.SMB.Username,
			Password: cfg.SMB.Password,
			Domain:   cfg.SMB.Domain,
		},
	}

	session, err := dialer.Dial(conn)
	if err != nil {
		logrus.Fatalf("SMB dial error: %v", err)
		return nil, err
	}
	defer session.Logoff()

	share_names, err := session.ListSharenames()
	if err != nil {
		logrus.Fatalf("Failed to enumerate shares: %v", err)
		return nil, err
	}

	shares := make([]Share, len(share_names))
	for i, n := range share_names {
		shares[i] = Share{ShareName: n}
	}
	return shares, nil
}
