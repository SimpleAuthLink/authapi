package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	api "github.com/simpleauthlink/authapi"
)

const (
	defaultHost         = "0.0.0.0"
	defaultPort         = 8080
	defaultDatabaseURI  = "mongodb://localhost:27017"
	defaultDatabaseName = "simpleauth"
	defaultEmailAddr    = ""
	defaultEmailPass    = ""
	defaultEmailHost    = ""
	defaultEmailPort    = 587

	hostFlag          = "host"
	portFlag          = "port"
	dbURIFlag         = "db-uri"
	dbNameFlag        = "db-name"
	emailAddrFlag     = "email-addr"
	emailPassFlag     = "email-pass"
	emailHostFlag     = "email-host"
	emailPortFlag     = "email-port"
	hostFlagDesc      = "service host"
	portFlagDesc      = "service port"
	dbURIFlagDesc     = "database uri"
	dbNameFlagDesc    = "database name"
	emailAddrFlagDesc = "email account address"
	emailPassFlagDesc = "email account password"
	emailHostFlagDesc = "email server host"
	emailPortFlagDesc = "email server port"

	hostEnv      = "SIMPLEAUTH_HOST"
	portEnv      = "SIMPLEAUTH_PORT"
	dbURIEnv     = "SIMPLEAUTH_DB_URI"
	dbNameEnv    = "SIMPLEAUTH_DB_NAME"
	emailAddrEnv = "SIMPLEAUTH_EMAIL_ADDR"
	emailPassEnv = "SIMPLEAUTH_EMAIL_PASS"
	emailHostEnv = "SIMPLEAUTH_EMAIL_HOST"
	emailPortEnv = "SIMPLEAUTH_EMAIL_PORT"
)

type config struct {
	host      string
	port      int
	dbURI     string
	dbName    string
	emailAddr string
	emailPass string
	emailHost string
	emailPort int
}

func main() {
	c, err := parseConfig()
	if err != nil {
		log.Fatalln("ERR: error parsing config:", err)
	}
	// create the service
	service, err := api.New(context.Background(), &api.Config{
		EmailConfig: api.EmailConfig{
			Address:   c.emailAddr,
			Password:  c.emailPass,
			EmailHost: c.emailHost,
			EmailPort: c.emailPort,
		},
		Server:          c.host,
		ServerPort:      c.port,
		DatabaseURI:     c.dbURI,
		DatabaseName:    c.dbName,
		CleanerCooldown: 30 * time.Minute,
	})
	if err != nil {
		log.Fatalln("ERR: error creating service:", err)
	}
	defer func() {
		if err := service.Stop(); err != nil {
			log.Fatalln("ERR: error stopping service:", err)
		}
	}()
	go func() {
		if err := service.Start(); err != nil {
			log.Fatalln("ERR: error running service:", err)
		}
	}()
	// wait for the service to finish
	service.WaitToShutdown()
}

func parseConfig() (*config, error) {
	var fhost, fdbURI, fdbName, femailAddr, femailPass, femailHost string
	var fport, femailPort int
	// get config from flags
	flag.StringVar(&fhost, hostFlag, defaultHost, hostFlagDesc)
	flag.IntVar(&fport, portFlag, defaultPort, hostFlagDesc)
	flag.StringVar(&fdbURI, dbURIFlag, defaultDatabaseURI, dbURIFlagDesc)
	flag.StringVar(&fdbName, dbNameFlag, defaultDatabaseName, dbNameFlagDesc)
	flag.StringVar(&femailAddr, emailAddrFlag, defaultEmailAddr, emailAddrFlagDesc)
	flag.StringVar(&femailPass, emailPassFlag, defaultEmailPass, emailPassFlagDesc)
	flag.StringVar(&femailHost, emailHostFlag, defaultEmailHost, emailHostFlagDesc)
	flag.IntVar(&femailPort, emailPortFlag, defaultEmailPort, emailPortFlagDesc)
	flag.Parse()
	// get config from env
	envHost := os.Getenv(hostEnv)
	envPort := os.Getenv(portEnv)
	envDBURI := os.Getenv(dbURIEnv)
	envDBName := os.Getenv(dbNameEnv)
	envEmailAddr := os.Getenv(emailAddrEnv)
	envEmailPass := os.Getenv(emailPassEnv)
	envEmailHost := os.Getenv(emailHostEnv)
	envEmailPort := os.Getenv(emailPortEnv)

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
		host:      fhost,
		port:      fport,
		dbURI:     fdbURI,
		dbName:    fdbName,
		emailAddr: femailAddr,
		emailPass: femailPass,
		emailHost: femailHost,
		emailPort: femailPort,
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
	if envDBURI != "" {
		c.dbURI = envDBURI
	}
	if envDBName != "" {
		c.dbName = envDBName
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
