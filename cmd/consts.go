package cmd

const (
	DefaultHost      = "0.0.0.0"
	DefaultPort      = 8080
	DefaultEmailAddr = ""
	DefaultEmailPass = ""
	DefaultEmailHost = ""
	DefaultEmailPort = 587
	DefaultSecret    = "simpleauthlink-secret"

	HostFlag          = "host"
	PortFlag          = "port"
	EmailAddrFlag     = "email-addr"
	EmailPassFlag     = "email-pass"
	EmailHostFlag     = "email-host"
	EmailPortFlag     = "email-port"
	SecretFlag        = "secret"
	HostFlagDesc      = "service host"
	PortFlagDesc      = "service port"
	EmailAddrFlagDesc = "email account address"
	EmailPassFlagDesc = "email account password"
	EmailHostFlagDesc = "email server host"
	EmailPortFlagDesc = "email server port"
	SecretFlagDesc    = "secret used to generate the tokens"

	HostEnv      = "HOST"
	PortEnv      = "PORT"
	EmailAddrEnv = "EMAIL_ADDR"
	EmailPassEnv = "EMAIL_PASS"
	EmailHostEnv = "EMAIL_HOST"
	EmailPortEnv = "EMAIL_PORT"
	SecretEnv    = "SECRET"
)
