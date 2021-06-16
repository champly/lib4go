package etcd3

type Option func(*Config)

type Config struct {
	// ServerList is the list of storage servers to connect with.
	ServerList []string
	// TLS credentials
	KeyFile       string
	CertFile      string
	TrustedCAFile string
	Prefix        string
}

func defaultConfig() *Config {
	return &Config{
		ServerList: []string{
			"127.0.0.1:2379",
		},
	}
}

func WithServerList(serverList []string) Option {
	return func(cfg *Config) {
		cfg.ServerList = serverList
	}
}

func WithTLSCredential(keyFile, certFile, trustedCAFile string) Option {
	return func(cfg *Config) {
		cfg.KeyFile = keyFile
		cfg.CertFile = certFile
		cfg.TrustedCAFile = trustedCAFile
	}
}
