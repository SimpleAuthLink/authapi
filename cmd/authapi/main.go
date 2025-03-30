package main

import (
	"context"
	"fmt"
	"log"

	"github.com/simpleauthlink/authapi/api"
	"github.com/simpleauthlink/authapi/cmd"
	"github.com/simpleauthlink/authapi/internal/osflag"
	"github.com/simpleauthlink/authapi/notification/email"
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

func (c *config) String() string {
	return fmt.Sprintf(`{"server": "%s:%d", "smtpServer": "%s:%d", "smtpAuth": "%s:%s", "secret": "%s"}`,
		c.host, c.port, c.emailHost, c.emailPort, c.emailAddr, c.emailPass, c.secret)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	c := new(config)
	// get config from flags
	osflag.StringVar(&c.host, cmd.HostEnv, cmd.HostFlag, cmd.DefaultHost, cmd.HostFlagDesc, false)
	osflag.IntVar(&c.port, cmd.PortEnv, cmd.PortFlag, cmd.DefaultPort, cmd.HostFlagDesc, false)
	osflag.StringVar(&c.emailAddr, cmd.EmailAddrEnv, cmd.EmailAddrFlag, cmd.DefaultEmailAddr, cmd.EmailAddrFlagDesc, true)
	osflag.StringVar(&c.emailPass, cmd.EmailPassEnv, cmd.EmailPassFlag, cmd.DefaultEmailPass, cmd.EmailPassFlagDesc, true)
	osflag.StringVar(&c.emailHost, cmd.EmailHostEnv, cmd.EmailHostFlag, cmd.DefaultEmailHost, cmd.EmailHostFlagDesc, true)
	osflag.IntVar(&c.emailPort, cmd.EmailPortEnv, cmd.EmailPortFlag, cmd.DefaultEmailPort, cmd.EmailPortFlagDesc, false)
	osflag.StringVar(&c.secret, cmd.SecretEnv, cmd.SecretFlag, cmd.DefaultSecret, cmd.SecretFlagDesc, true)
	if err := osflag.Parse(); err != nil {
		log.Fatalln("ERR: error parsing flags:", err)
	}
	if !osflag.Parsed() {
		log.Fatalln("ERR: error parsing flags:", "flags not parsed")
		osflag.PrintDefaults()
	}
	log.Println("INF: starting service with config:", c.String())
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
	// start the service in background
	go func() {
		if err := service.Start(); err != nil {
			log.Fatalln("ERR: error running service:", err)
		}
	}()
	// wait for the service to finish
	if err := service.WaitToShutdown(); err != nil {
		log.Fatalln("ERR: error waiting for service to finish:", err)
	}
}
