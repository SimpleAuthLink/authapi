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

	HostEnv      = "SIMPLEAUTH_HOST"
	PortEnv      = "SIMPLEAUTH_PORT"
	EmailAddrEnv = "SIMPLEAUTH_EMAIL_ADDR"
	EmailPassEnv = "SIMPLEAUTH_EMAIL_PASS"
	EmailHostEnv = "SIMPLEAUTH_EMAIL_HOST"
	EmailPortEnv = "SIMPLEAUTH_EMAIL_PORT"
	SecretEnv    = "SIMPLEAUTH_SECRET"
)
