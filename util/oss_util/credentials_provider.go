package oss_util

import (
	"context"
	"fmt"
	"os"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

type XCredentialsProvider struct {
	accessKey    string
	accessSecret string
}

func (s *XCredentialsProvider) GetCredentials(ctx context.Context) (credentials.Credentials, error) {
	id := s.accessKey
	secret := s.accessSecret
	if id == "" || secret == "" {
		return credentials.Credentials{}, fmt.Errorf("access key id or access key secret is empty!")
	}
	return credentials.Credentials{
		AccessKeyID:     id,
		AccessKeySecret: secret,
		SecurityToken:   os.Getenv("OSS_SESSION_TOKEN"),
	}, nil
}

func NewXCredentialsProvider(accessKey, accessSecret string) credentials.CredentialsProvider {
	return &XCredentialsProvider{
		accessKey:    accessKey,
		accessSecret: accessSecret,
	}
}
