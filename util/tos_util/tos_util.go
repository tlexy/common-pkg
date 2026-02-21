package tos_util

import (
	"encoding/json"
	"fmt"

	"github.com/volcengine/volc-sdk-golang/service/sts"
)

func GetTemporaryCredentials(accessKey, secretKey, endPoint, bucketName, tosTrn string, sec int64) (string, string, string, error) {
	//参考： https://www.volcengine.com/docs/6349/127695?lang=zh
	sts.DefaultInstance.Client.SetAccessKey(accessKey)
	sts.DefaultInstance.Client.SetSecretKey(secretKey)

	list, status, err := sts.DefaultInstance.AssumeRole(&sts.AssumeRoleRequest{
		DurationSeconds: int(sec),
		RoleTrn:        tosTrn,
		RoleSessionName: "jest_for_test",
	})
	fmt.Println(status, err)
	b, _ := json.Marshal(list)
	fmt.Println(string(b))

	if list.Result == nil {
		return "", "", "", fmt.Errorf("assume role failed, list.Result == nil")
	}
	if list.Result.Credentials == nil {
		return "", "", "", fmt.Errorf("assume role failed, list.Result.Credentials == nil")
	}
	return list.Result.Credentials.AccessKeyId, list.Result.Credentials.SecretAccessKey, list.Result.Credentials.SessionToken, nil
}