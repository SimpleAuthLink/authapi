package main

import (
	"context"
	"log"

	"github.com/simpleauthlink/authapi/api"
	"github.com/simpleauthlink/authapi/cmd"
	"github.com/simpleauthlink/authapi/internal/osflag"
	"github.com/simpleauthlink/authapi/notification/email"
)

func main() {
	var (
		demoServer string
		demoPort   int
		demoSecret string
	)
	osflag.StringVar(&demoServer, cmd.HostEnv, cmd.HostFlag, cmd.DefaultHost, cmd.HostFlagDesc, false)
	osflag.IntVar(&demoPort, cmd.PortEnv, cmd.PortFlag, cmd.DefaultPort, cmd.PortFlagDesc, false)
	osflag.StringVar(&demoSecret, cmd.SecretEnv, cmd.SecretFlag, cmd.DefaultSecret, cmd.SecretFlagDesc, false)
	if err := osflag.Parse(nil); err != nil {
		log.Fatalln("ERR: error parsing flags:", err)
	}
	log.Println("INF: starting service with config:", demoServer, demoPort, demoSecret)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// create the email queue
	emailQueue, err := email.NewEmailQueue(context.Background(), &email.EmailConfig{
		FromName:    "SimpleAuthLink Demo",
		FromAddress: "demo@simpleauth.link",
		SMTPServer:  demoServer,
		SMTPPort:    2525,
	})
	if err != nil {
		log.Fatalln("WRN: something occurs during email queue creation:", err)
	}
	// start the email queue and defer to stop it
	emailQueue.Start()
	defer emailQueue.Stop()
	// create the service
	service, err := api.New(ctx, &api.Config{
		Server:       demoServer,
		ServerPort:   demoPort,
		Secret:       demoSecret,
		DemoMode:     true,
		DemoSMTPAddr: demoServer,
		DemoSMTPPort: 2525,
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
