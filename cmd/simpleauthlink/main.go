package main

import (
	"flag"
	"log"

	"github.com/lucasmenendez/authapi"
)

func main() {
	serviceHost := flag.String("host", "0.0.0.0", "service host")
	servicePort := flag.Int("port", 8080, "service port")
	dataPath := flag.String("data-path", "./.temp", "data path")
	emailAddr := flag.String("email-addr", "", "email account address")
	emailPass := flag.String("email-pass", "", "email account password")
	emailHost := flag.String("email-host", "", "email server host")
	emailPort := flag.Int("email-port", 587, "email server port")
	flag.Parse()
	// check if the required flags are set
	if len(*emailAddr) == 0 {
		log.Fatal("ERR: email address is required, use -email-addr")
	}
	if len(*emailPass) == 0 {
		log.Fatal("ERR: email password is required, use -email-pass")
	}
	if len(*emailHost) == 0 {
		log.Fatal("ERR: email host is required, use -email-host")
	}
	// create the service
	service, err := authapi.New(&authapi.Config{
		EmailConfig: authapi.EmailConfig{
			Address:   *emailAddr,
			Password:  *emailPass,
			EmailHost: *emailHost,
			EmailPort: *emailPort,
		},
		Server:     *serviceHost,
		ServerPort: *servicePort,
		DataPath:   *dataPath,
	})
	if err != nil {
		log.Fatalln("ERR: error creating service:", err)
	}
	defer func() {
		if err := service.Stop(); err != nil {
			log.Fatalln("ERR: error stopping service:", err)
		}
	}()
	if err := service.Start(); err != nil {
		log.Fatalln("ERR: error running service:", err)
	}
}
