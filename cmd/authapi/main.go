package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/simpleauthlink/authapi/api"
	"github.com/simpleauthlink/authapi/email"
)

const (
	defaultHost               = "0.0.0.0"
	defaultPort               = 8080
	defaultEmailAddr          = ""
	defaultEmailPass          = ""
	defaultEmailHost          = ""
	defaultEmailPort          = 587
	defaultTokenEmailTemplate = "assets/token_email_template.html"
	defaultAppEmailTemplate   = "assets/app_email_template.html"

	hostFlag               = "host"
	portFlag               = "port"
	emailAddrFlag          = "email-addr"
	emailPassFlag          = "email-pass"
	emailHostFlag          = "email-host"
	emailPortFlag          = "email-port"
	tokenEmailTemplateFlag = "email-token-template"
	appEmailTemplateFlag   = "email-app-template"
	disposableSrcFlag      = "disposable-src"
	hostFlagDesc           = "service host"
	portFlagDesc           = "service port"
	dbURIFlagDesc          = "database uri"
	dbNameFlagDesc         = "database name"
	emailAddrFlagDesc      = "email account address"
	emailPassFlagDesc      = "email account password"
	emailHostFlagDesc      = "email server host"
	emailPortFlagDesc      = "email server port"
	tokenEmailTemplateDesc = "path to the html template of new token email"
	appEmailTemplateDesc   = "path to the html template of new app email"

	hostEnv               = "SIMPLEAUTH_HOST"
	portEnv               = "SIMPLEAUTH_PORT"
	emailAddrEnv          = "SIMPLEAUTH_EMAIL_ADDR"
	emailPassEnv          = "SIMPLEAUTH_EMAIL_PASS"
	emailHostEnv          = "SIMPLEAUTH_EMAIL_HOST"
	emailPortEnv          = "SIMPLEAUTH_EMAIL_PORT"
	tokenEmailTemplateEnv = "SIMPLEAUTH_TOKEN_EMAIL_TEMPLATE"
	appEmailTemplateEnv   = "SIMPLEAUTH_APP_EMAIL_TEMPLATE"
)

type config struct {
	host               string
	port               int
	dbURI              string
	dbName             string
	emailAddr          string
	emailPass          string
	emailHost          string
	emailPort          int
	tokenEmailTemplate string
	appEmailTemplate   string
	disposableSrc      string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	c, err := parseConfig()
	if err != nil {
		log.Fatalln("ERR: error parsing config:", err)
	}
	// create the service
	service, err := api.New(context.Background(), &api.Config{
		EmailConfig: email.EmailConfig{
			FromName:     "SimpleAuthLink",
			FromAddress:  c.emailAddr,
			SMTPUsername: c.emailAddr,
			SMTPPassword: c.emailPass,
			SMTPServer:   c.emailHost,
			SMTPPort:     c.emailPort,
		},
		Server:          c.host,
		ServerPort:      c.port,
		CleanerCooldown: 30 * time.Minute,
	})
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
	var fhost, fdbURI, fdbName, femailAddr, femailPass, femailHost, ftokenEmailTemplate, fappEmailTemplate, fdisposableSrc string
	var fport, femailPort int
	// get config from flags
	flag.StringVar(&fhost, hostFlag, defaultHost, hostFlagDesc)
	flag.IntVar(&fport, portFlag, defaultPort, hostFlagDesc)
	flag.StringVar(&femailAddr, emailAddrFlag, defaultEmailAddr, emailAddrFlagDesc)
	flag.StringVar(&femailPass, emailPassFlag, defaultEmailPass, emailPassFlagDesc)
	flag.StringVar(&femailHost, emailHostFlag, defaultEmailHost, emailHostFlagDesc)
	flag.StringVar(&ftokenEmailTemplate, tokenEmailTemplateFlag, defaultTokenEmailTemplate, tokenEmailTemplateDesc)
	flag.StringVar(&fappEmailTemplate, appEmailTemplateFlag, defaultAppEmailTemplate, appEmailTemplateDesc)
	flag.IntVar(&femailPort, emailPortFlag, defaultEmailPort, emailPortFlagDesc)
	flag.Parse()
	// get config from env
	envHost := os.Getenv(hostEnv)
	envPort := os.Getenv(portEnv)
	envEmailAddr := os.Getenv(emailAddrEnv)
	envEmailPass := os.Getenv(emailPassEnv)
	envEmailHost := os.Getenv(emailHostEnv)
	envEmailPort := os.Getenv(emailPortEnv)
	envtokenEmailTemplate := os.Getenv(tokenEmailTemplateEnv)
	envAppEmailTemplate := os.Getenv(appEmailTemplateEnv)

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
	// set flags values by default
	c := &config{
		host:               fhost,
		port:               fport,
		dbURI:              fdbURI,
		dbName:             fdbName,
		emailAddr:          femailAddr,
		emailPass:          femailPass,
		emailHost:          femailHost,
		emailPort:          femailPort,
		tokenEmailTemplate: ftokenEmailTemplate,
		appEmailTemplate:   fappEmailTemplate,
		disposableSrc:      fdisposableSrc,
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
	if envtokenEmailTemplate != "" {
		c.tokenEmailTemplate = envtokenEmailTemplate
	}
	if envAppEmailTemplate != "" {
		c.appEmailTemplate = envAppEmailTemplate
	}
	return c, nil
}
