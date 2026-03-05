package vod_volce

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/tlexy/common-pkg/util/volce_base"
)

// https://www.volcengine.com/docs/4/1923688?lang=zh
// func (v *VodVolce) SubmitOcrTask(vid string) error {
// 	region := "cn-north-1"
// 	config := volcengine.NewConfig().
// 		WithCredentials(credentials.NewStaticCredentials(v.accessKey, v.secretKey, "")).
// 		WithRegion(region)

// 	sess, err := session.NewSession(config)
// 	if err != nil {
// 		panic(err)
// 	}

// 	client := vod20250101.New(sess)

// 	resp, err := client.StartExecution(&vod20250101.StartExecutionInput{
// 		Input: &vod20250101.InputForStartExecutionInput{
// 			Type: volcengine.String("Vid"),
// 			Vid:  volcengine.String(vid),
// 		},
// 		Operation: &vod20250101.ConvertOperationForStartExecutionInput{
// 			Type: volcengine.String("Task"),
// 			Task: &vod20250101.TaskForStartExecutionInput{
// 				Type:      volcengine.String("Highlight"),
// 				Highlight: &vod20250101.HighlightForStartExecutionInput{},
// 			},
// 		},
// 		Control: &vod20250101.ControlForStartExecutionInput{
// 			ClientToken: volcengine.String("testToken"),
// 		},
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println(resp)
// 	return nil
// }

func (v *VodVolce) SubmitOcrTask(vid, spaceName string) error {
	vodv2Session := volce_base.NewVodV2Session(v.accessKey, v.secretKey)

	startExecInput := &volce_base.StartExecutionInput{
		Input: &volce_base.InputForStartExecutionInput{
			Type: "Vid",
			Vid:  vid,
		},
		Operation: &volce_base.ConvertOperationForStartExecutionInput{
			Type: "Task",
			Task: &volce_base.Task{
				Type: "Ocr",
				Ocr: &volce_base.OperationTaskOcr{
					WithImageSet: false,
					Mode:         "Detailed",
				},
			},
		},
		SpaceName: spaceName,
	}

	body, err := json.Marshal(startExecInput)
	if err != nil {
		return err
	}

	req, err := vodv2Session.ConstructHttpRequest(body)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request err: %w", err)
	}

	// 5. 打印响应
	// responseRaw, err := httputil.DumpResponse(response, true)
	// if err != nil {
	// 	return fmt.Errorf("dump response err: %w", err)
	// }

	// log.Printf("response:\n%s\n", string(responseRaw))

	if response.StatusCode == 200 {
		log.Printf("请求成功")
	} else {
		log.Printf("请求失败")
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("read response body err: %w", err)
		}
		log.Printf("Response Body: %s", string(responseBody))
		return fmt.Errorf("request failed, status: %s, response body: %s", response.Status, string(responseBody))
	}
	return nil
}
