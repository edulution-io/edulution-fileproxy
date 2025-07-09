package types

type SmbCreds struct {
	User   string
	Pass   string
	Domain string
}

func NewSmbCreds(user string, pass string, domain string) *SmbCreds {
	return &SmbCreds{
		User:   user,
		Pass:   pass,
		Domain: domain,
	}
}
