package main

import (
	"context"
	"os"

	"github.com/simpleauthlink/authapi/api"
	"github.com/simpleauthlink/authapi/notification"
	"github.com/simpleauthlink/authapi/notification/email"
	"go.k7z7z.cc/x/flag"
	"go.k7z7z.cc/x/log"
	"go.k7z7z.cc/x/proc"
)

type Config struct {
	email.EmailConfig
	APIHost                  string `env:"HOST" flag:"host" sflag:"h" usage:"service host"`
	APIPort                  int    `env:"PORT" flag:"port" sflag:"p" usage:"service port"`
	Secret                   string `env:"SECRET" flag:"secret" sflag:"s" usage:"secret used to generate the tokens"`
	NotificationQueueSize    int    `env:"NOTIFICATION_QUEUE_SIZE" flag:"queue-size" usage:"size of the notification queue"`
	NotificationQueueWorkers int    `env:"NOTIFICATION_QUEUE_WORKERS" flag:"queue-workers" usage:"number of workers of the notification queue"`
}

func main() {
	log.SetLevel(log.DebugLevel)
	c := &Config{
		APIHost: "0.0.0.0",
		APIPort: 8080,
		EmailConfig: email.EmailConfig{
			SMTPPort: 587,
		},
		NotificationQueueSize:    1000,
		NotificationQueueWorkers: 10,
	}
	envvars, err := proc.EnvvarsFromFile(".env")
	if err != nil {
		log.Errw("error loading env file", "err", err)
		return
	}
	if err := proc.UnmarshalEnvvars(envvars, c); err != nil {
		log.Errw("error processing envvars", "err", err)
		return
	}
	if err := flag.UnmarshalArgs(os.Args[1:], c); err != nil {
		log.Errw("error loading flags", "err", err)
		return
	}

	log.Infow("starting service with config:",
		"api-host", c.APIHost,
		"api-port", c.APIPort,
		"queue-size", c.NotificationQueueSize,
		"queue-workers", c.NotificationQueueWorkers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// create notification queue
	notificationQueue, err := notification.NewQueue(ctx, c.NotificationQueueSize, &c.EmailConfig)
	if err != nil {
		log.Errw("error creating notification queue", "err", err)
		return
	}
	// start the email queue and defer to stop it
	notificationQueue.Start(c.NotificationQueueWorkers)
	defer notificationQueue.Stop()
	// create the apiService
	apiService, err := api.New(ctx, &api.Config{
		Server:     c.APIHost,
		ServerPort: c.APIPort,
		Secret:     c.Secret,
	}, notificationQueue)
	if err != nil {
		log.Errw("error creating service", "err", err)
		return
	}
	// start the service and wait for it to finish
	if err = new(proc.Shutdown).WithInterrupt().WithRecovery().WithRun(func(ctx context.Context) error {
		return apiService.Start()
	}).WithCallback(func(_ context.Context) error {
		return apiService.Stop()
	}).Wait(ctx); err != nil {
		log.Errw("error during shutdown", "err", err)
		return
	}
	log.Info("HTTP server stopped, exiting...")
}
