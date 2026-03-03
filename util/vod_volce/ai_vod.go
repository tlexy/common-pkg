package vod_volce

import (
	"fmt"

	"github.com/volcengine/volcengine-go-sdk/service/vod20250101"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/session"
)

// https://www.volcengine.com/docs/4/1923688?lang=zh
func (v *VodVolce) SubmitOcrTask(vid string) error {
	region := "cn-north-1"
	config := volcengine.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(v.accessKey, v.secretKey, "")).
		WithRegion(region)

	sess, err := session.NewSession(config)
	if err != nil {
		panic(err)
	}

	client := vod20250101.New(sess)

	resp, err := client.StartExecution(&vod20250101.StartExecutionInput{
		Input: &vod20250101.InputForStartExecutionInput{
			Type: volcengine.String("Vid"),
			Vid:  volcengine.String(vid),
		},
		Operation: &vod20250101.ConvertOperationForStartExecutionInput{
			Type: volcengine.String("Task"),
			Task: &vod20250101.TaskForStartExecutionInput{
				Type:      volcengine.String("Highlight"),
				Highlight: &vod20250101.HighlightForStartExecutionInput{},
			},
		},
		Control: &vod20250101.ControlForStartExecutionInput{
			ClientToken: volcengine.String("testToken"),
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)
	return nil
}
