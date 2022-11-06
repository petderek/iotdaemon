package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os/user"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/petderek/sqsbuddy"
)

const (
	federateCmd = "/bin/federator creds -region %s --role-arn %s --json"
)

var (
	address    = flag.String("address", "", "the full ssh address, eg user@domain:port")
	keypath    = flag.String("keypath", "", "the path to the id_rsa key")
	keypass    = flag.String("keypass", "", "ssh passphrase to use")
	region     = flag.String("region", "", "the aws region to use")
	role       = flag.String("role-arn", "", "the role to assume")
	queue      = flag.String("queue", "", "the sqs queue url to use")
	serialPort = flag.String("serial", "", "the serial port name")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	if address == nil || *address == "" {
		log.Fatalln("address is required")
	}
	urldata, err := url.Parse(*address)
	if err != nil {
		log.Fatalln("address is not a uri: ", err)
	}

	who := urldata.User.Username()
	if who == "" {
		if u, err := user.Current(); err == nil {
			who = u.Username
		} else {
			log.Fatalln("user not supplied and cannot be inferred: ", err)
		}
	}

	ssh := &sqsbuddy.SSHBuddy{
		User:          who,
		Address:       urldata.Host,
		Command:       fmt.Sprintf(federateCmd, *region, *role),
		KeyPath:       *keypath,
		KeyPassphrase: *keypass,
		InsecureHosts: true,
	}

	provider := &sqsbuddy.CredsBuddy{SSH: ssh}
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(*region),
		config.WithCredentialsProvider(provider),
	)
	if err != nil {
		log.Fatalln("unable to load config: ", err)
	}

	buddy := sqsbuddy.SQSBuddy{
		Config:  cfg,
		Context: ctx,
		Url:     *queue,
	}

	queue := buddy.Poll()

	for msg, ok := <-queue; ok; msg, ok = <-queue {
		if msg != nil {
			log.Println("message: ", *msg)
		}
	}
	log.Fatalln("channel closed")
}
