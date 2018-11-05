package session

import (
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	"github.com/whosonfirst/go-whosonfirst-aws/config"
)

func NewSessionWithCredentials(str_creds string, region string) (*aws_session.Session, error) {

	cfg, err := config.NewConfigWithCredentials(str_creds, region)

	if err != nil {
		return nil, err
	}

	sess := aws_session.New(cfg)

	_, err = sess.Config.Credentials.Get()

	if err != nil {
		return nil, err
	}

	return sess, nil
}
