package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/simpleauthlink/authapi/api"
	"github.com/simpleauthlink/authapi/notification/email"
)

const (
	defaultHost      = "0.0.0.0"
	defaultPort      = 8080
	defaultEmailAddr = ""
	defaultEmailPass = ""
	defaultEmailHost = ""
	defaultEmailPort = 587
	defaultSecret    = "simpleauthlink-secret"

	hostFlag          = "host"
	portFlag          = "port"
	emailAddrFlag     = "email-addr"
	emailPassFlag     = "email-pass"
	emailHostFlag     = "email-host"
	emailPortFlag     = "email-port"
	secretFlag        = "secret"
	hostFlagDesc      = "service host"
	portFlagDesc      = "service port"
	emailAddrFlagDesc = "email account address"
	emailPassFlagDesc = "email account password"
	emailHostFlagDesc = "email server host"
	emailPortFlagDesc = "email server port"
	secretFlagDesc    = "secret used to generate the tokens"

	hostEnv      = "SIMPLEAUTH_HOST"
	portEnv      = "SIMPLEAUTH_PORT"
	emailAddrEnv = "SIMPLEAUTH_EMAIL_ADDR"
	emailPassEnv = "SIMPLEAUTH_EMAIL_PASS"
	emailHostEnv = "SIMPLEAUTH_EMAIL_HOST"
	emailPortEnv = "SIMPLEAUTH_EMAIL_PORT"
	secretEnv    = "SIMPLEAUTH_SECRET"
)

type config struct {
	host      string
	port      int
	emailAddr string
	emailPass string
	emailHost string
	emailPort int
	secret    string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	c, err := parseConfig()
	if err != nil {
		log.Fatalln("ERR: error parsing config:", err)
	}
	// create email queue
	emailQueue, err := email.NewEmailQueue(context.Background(), &email.EmailConfig{
		FromName:     "SimpleAuthLink",
		FromAddress:  c.emailAddr,
		SMTPUsername: c.emailAddr,
		SMTPPassword: c.emailPass,
		SMTPServer:   c.emailHost,
		SMTPPort:     c.emailPort,
	})
	if err != nil {
		log.Fatalln("WRN: something occurs during email queue creation:", err)
	}
	// start the email queue and defer to stop it
	emailQueue.Start()
	defer emailQueue.Stop()
	// create the service
	service, err := api.New(context.Background(), &api.Config{
		Server:     c.host,
		ServerPort: c.port,
		Secret:     c.secret,
	}, emailQueue)
	if err != nil {
		log.Fatalln("ERR: error creating service:", err)
	}
	go func() {
		if err := service.Start(); err != nil {
			log.Fatalln("ERR: error running service:", err)
		}
	}()
	// wait for the service to finish
	service.WaitToShutdown()
}

func parseConfig() (*config, error) {
	var fhost, femailAddr, femailPass, femailHost, fsecret string
	var fport, femailPort int
	// get config from flags
	flag.StringVar(&fhost, hostFlag, defaultHost, hostFlagDesc)
	flag.IntVar(&fport, portFlag, defaultPort, hostFlagDesc)
	flag.StringVar(&femailAddr, emailAddrFlag, defaultEmailAddr, emailAddrFlagDesc)
	flag.StringVar(&femailPass, emailPassFlag, defaultEmailPass, emailPassFlagDesc)
	flag.StringVar(&femailHost, emailHostFlag, defaultEmailHost, emailHostFlagDesc)
	flag.IntVar(&femailPort, emailPortFlag, defaultEmailPort, emailPortFlagDesc)
	flag.StringVar(&fsecret, secretFlag, defaultSecret, secretFlagDesc)
	flag.Parse()
	// get config from env
	envHost := os.Getenv(hostEnv)
	envPort := os.Getenv(portEnv)
	envEmailAddr := os.Getenv(emailAddrEnv)
	envEmailPass := os.Getenv(emailPassEnv)
	envEmailHost := os.Getenv(emailHostEnv)
	envEmailPort := os.Getenv(emailPortEnv)
	envSecret := os.Getenv(secretEnv)
	// check if the required flags are set
	if femailAddr == "" && envEmailAddr == "" {
		return nil, fmt.Errorf("email address is required, use -%s or set %s env var", emailAddrFlag, emailAddrEnv)
	}
	if femailPass == "" && envEmailPass == "" {
		return nil, fmt.Errorf("email password is required, use -%s or set %s env var", emailPassFlag, emailPassEnv)
	}
	if femailHost == "" && envEmailHost == "" {
		return nil, fmt.Errorf("email host is required, use -%s or set %s env var", emailHostFlag, emailHostEnv)
	}
	if fsecret == "" && envSecret == "" {
		return nil, fmt.Errorf("secret is required, use -%s or set %s env var", secretFlag, secretEnv)
	}
	// set flags values by default
	c := &config{
		host:      fhost,
		port:      fport,
		emailAddr: femailAddr,
		emailPass: femailPass,
		emailHost: femailHost,
		emailPort: femailPort,
		secret:    fsecret,
	}
	// if some flags are not set, set them by env
	if envHost != "" {
		c.host = envHost
	}
	if envPort != "" {
		if nenvPort, err := strconv.Atoi(envPort); err == nil {
			c.port = nenvPort
		} else {
			return nil, fmt.Errorf("invalid port value: %s", envPort)
		}
	}
	if envEmailAddr != "" {
		c.emailAddr = envEmailAddr
	}
	if envEmailPass != "" {
		c.emailPass = envEmailPass
	}
	if envEmailHost != "" {
		c.emailHost = envEmailHost
	}
	if envEmailPort != "" {
		if nenvEmailPort, err := strconv.Atoi(envEmailPort); err == nil {
			c.emailPort = nenvEmailPort
		} else {
			return nil, fmt.Errorf("invalid email port value: %s", envEmailPort)
		}
	}
	return c, nil
}
