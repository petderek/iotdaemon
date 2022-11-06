package sqsbuddy

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	maxMessages = 1
	longPoll    = 20
)

type SQSBuddy struct {
	Config  aws.Config
	Context context.Context
	Url     string
	Logger  *log.Logger
	_queue  chan *string
	_init   sync.Once
}

// Poll runs a background thread and continuously grabs strings from SQS.
func (s *SQSBuddy) Poll() <-chan *string {
	s.doInit()
	return s._queue
}

func (s *SQSBuddy) doPoll() {
	client := sqs.NewFromConfig(s.Config)
	req := &sqs.ReceiveMessageInput{
		QueueUrl:            &s.Url,
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     longPoll,
	}

	for {
		select {
		case <-s.Context.Done():
			close(s._queue)
			return

		default:
			res, err := client.ReceiveMessage(s.Context, req)
			if err != nil {
				s.Logger.Println("error polling: ", err.Error())
				time.Sleep(30 * time.Second)
				continue
			}

			for _, message := range res.Messages {
				s._queue <- message.Body
				_, err = client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
					QueueUrl:      &s.Url,
					ReceiptHandle: message.ReceiptHandle,
				})
				if err != nil {
					s.Logger.Println("error deleting: ", err.Error())
				}
			}
			// empty, retry
		}
	}
}

func (s *SQSBuddy) doInit() {
	s._init.Do(func() {
		if s.Context == nil {
			s.Context = context.Background()
		}
		if s.Logger == nil {
			s.Logger = log.Default()
		}

		s._queue = make(chan *string, 1)
		if s.Url == "" {
			close(s._queue)
			return
		}
		go s.doPoll()
	})
}
