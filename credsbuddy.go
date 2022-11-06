package iotdaemon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type CredsBuddy struct {
	SSH *SSHBuddy
}

func (cb *CredsBuddy) Retrieve(ctx context.Context) (aws.Credentials, error) {
	var creds aws.Credentials
	data, err := cb.SSH.Run()
	if err != nil {
		return creds, err
	}
	var read map[string]string
	if err = json.Unmarshal(data, &read); err != nil {
		return creds, err
	}

	parsedTime, err := time.Parse(time.RFC3339, read["expiresAfter"])
	if err != nil {
		return creds, err
	}

	creds.CanExpire = true
	creds.AccessKeyID = read["AccessKeyId"]
	creds.SecretAccessKey = read["SecretAccessKey"]
	creds.SessionToken = read["SessionToken"]
	creds.Expires = parsedTime

	return creds, nil
}
