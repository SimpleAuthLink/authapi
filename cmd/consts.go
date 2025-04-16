package cmd

const (
	DefaultHost      = "0.0.0.0"
	DefaultPort      = 8080
	DefaultEmailAddr = ""
	DefaultEmailUser = ""
	DefaultEmailPass = ""
	DefaultEmailHost = ""
	DefaultEmailPort = 587
	DefaultSecret    = "simpleauthlink-secret"

	HostFlag          = "host"
	PortFlag          = "port"
	EmailAddrFlag     = "email-addr"
	EmailUserFlag     = "email-user"
	EmailPassFlag     = "email-pass"
	EmailHostFlag     = "email-host"
	EmailPortFlag     = "email-port"
	SecretFlag        = "secret"
	HostFlagDesc      = "service host"
	PortFlagDesc      = "service port"
	EmailAddrFlagDesc = "email account address"
	EmailUserFlagDesc = "email account username"
	EmailPassFlagDesc = "email account password"
	EmailHostFlagDesc = "email server host"
	EmailPortFlagDesc = "email server port"
	SecretFlagDesc    = "secret used to generate the tokens"

	HostEnv      = "HOST"
	PortEnv      = "PORT"
	EmailAddrEnv = "EMAIL_ADDR"
	EmailUserEnv = "EMAIL_USER"
	EmailPassEnv = "EMAIL_PASS"
	EmailHostEnv = "EMAIL_HOST"
	EmailPortEnv = "EMAIL_PORT"
	SecretEnv    = "SECRET"
)
