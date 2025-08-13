package pulsar

type ClientConfig struct {
	ServiceURL string      `yaml:"serviceUrl"`
	Auth       *AuthConfig `yaml:"auth,omitempty"`
}

type AuthConfig struct {
	Type  string `yaml:"type"`
	Token string `yaml:"token,omitempty"`
	Cert  string `yaml:"cert,omitempty"`
	Key   string `yaml:"key,omitempty"`
}

func setDefaults(config *ClientConfig) {
}
