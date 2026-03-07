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

func (v *VodVolce) SubmitOcrTask(vid, spaceName string) (string, error) {
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
		return "", fmt.Errorf("marshal start exec input err: %w", err)
	}

	req, err := vodv2Session.StartExecutionRequest(body)
	if err != nil {
		return "", fmt.Errorf("start execution request err: %w", err)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request err: %w", err)
	}

	// 5. 打印响应
	// responseRaw, err := httputil.DumpResponse(response, true)
	// if err != nil {
	// 	return fmt.Errorf("dump response err: %w", err)
	// }

	// log.Printf("response:\n%s\n", string(responseRaw))

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read response body err: %w", err)
	}
	//log.Printf("Response Body: %s", string(responseBody))
	if response.StatusCode != 200 {
		//log.Printf("请求失败")
		return "", fmt.Errorf("request failed, status code: %d, body: %s", response.StatusCode, string(responseBody))
	}
	//log.Printf("请求成功")
	// 解析 JSON 响应
	var startExecOutput volce_base.StartExecutionOutput
	err = json.Unmarshal(responseBody, &startExecOutput)
	if err != nil {
		return "", fmt.Errorf("unmarshal response body err: %w", err)
	}

	return startExecOutput.Result.RunId, nil
}

func (v *VodVolce) QueryOcrTaskResult(runId string) (*volce_base.ExecutionOcrResult, error) {
	vodv2Session := volce_base.NewVodV2Session(v.accessKey, v.secretKey)

	req, err := vodv2Session.GetExecutionRequest(runId)
	if err != nil {
		return nil, fmt.Errorf("get execution request err: %w", err)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request err: %w", err)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body err: %w", err)
	}
	//log.Printf("Response Body: %s", string(responseBody))
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("request failed, status code: %d, body: %s", response.StatusCode, string(responseBody))
	}

	var execOutput volce_base.ExecutionOcrResult
	err = json.Unmarshal(responseBody, &execOutput)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response body err: %w", err)
	}
	return &execOutput, nil
}

func (v *VodVolce) SubmitEraseTask(vid, spaceName string) (string, error) {
	vodv2Session := volce_base.NewVodV2Session(v.accessKey, v.secretKey)

	startExecInput := &volce_base.StartExecutionInput{
		Input: &volce_base.InputForStartExecutionInput{
			Type: "Vid",
			Vid:  vid,
		},
		Operation: &volce_base.ConvertOperationForStartExecutionInput{
			Type: "Task",
			Task: &volce_base.Task{
				Type: "Erase",
				Erase: &volce_base.OperationTaskErase{
					Mode: "Auto",
					Auto: &volce_base.EraseAuto{
						Type: "Subtitle",
					},
					NewVid:        true,
					WithEraseInfo: true,
				},
			},
		},
		SpaceName: spaceName,
	}

	body, err := json.Marshal(startExecInput)
	if err != nil {
		return "", fmt.Errorf("marshal start exec input err: %w", err)
	}

	req, err := vodv2Session.StartExecutionRequest(body)
	if err != nil {
		return "", fmt.Errorf("start execution request err: %w", err)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request err: %w", err)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read response body err: %w", err)
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("request failed, status code: %d, body: %s", response.StatusCode, string(responseBody))
	}

	var startExecOutput volce_base.StartExecutionOutput
	err = json.Unmarshal(responseBody, &startExecOutput)
	if err != nil {
		return "", fmt.Errorf("unmarshal response body err: %w", err)
	}

	return startExecOutput.Result.RunId, nil
}

func (v *VodVolce) QueryEraseTaskResult(runId string) (*volce_base.ExecutionEraseResult, error) {
	vodv2Session := volce_base.NewVodV2Session(v.accessKey, v.secretKey)

	req, err := vodv2Session.GetExecutionRequest(runId)
	if err != nil {
		return nil, fmt.Errorf("get execution request err: %w", err)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request err: %w", err)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body err: %w", err)
	}
	log.Printf("Response Body: %s", string(responseBody))
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("request failed, status code: %d, body: %s", response.StatusCode, string(responseBody))
	}

	var execOutput volce_base.ExecutionEraseResult
	err = json.Unmarshal(responseBody, &execOutput)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response body err: %w", err)
	}
	return &execOutput, nil
}

func (v *VodVolce) SubmitAudioExtractTask(vid, spaceName string) (string, error) {
	vodv2Session := volce_base.NewVodV2Session(v.accessKey, v.secretKey)

	startExecInput := &volce_base.StartExecutionInput{
		Input: &volce_base.InputForStartExecutionInput{
			Type: "Vid",
			Vid:  vid,
		},
		Operation: &volce_base.ConvertOperationForStartExecutionInput{
			Type: "Task",
			Task: &volce_base.Task{
				Type: "AudioExtract",
				AudioExtract: &volce_base.OperationTaskAudioExtract{
					Voice: true,
				},
			},
		},
		SpaceName: spaceName,
	}

	body, statusCode, err := vodv2Session.StartExecution(startExecInput)
	if err != nil {
		return "", fmt.Errorf("start execution err: %w", err)
	}
	log.Printf("StartExecution status code: %d", statusCode)
	if statusCode != 200 {
		return "", fmt.Errorf("request failed, status code: %d, body: %s", statusCode, string(body))
	}

	var startExecOutput volce_base.StartExecutionOutput
	err = json.Unmarshal(body, &startExecOutput)
	if err != nil {
		return "", fmt.Errorf("unmarshal response body err: %w", err)
	}
	return startExecOutput.Result.RunId, nil
}

func (v *VodVolce) QueryAudioExtractTaskResult(runId string) (*volce_base.ExecutionAudioExtractResult, error) {
	vodv2Session := volce_base.NewVodV2Session(v.accessKey, v.secretKey)

	body, statusCode, err := vodv2Session.GetExecutionResult(runId)
	if err != nil {
		return nil, fmt.Errorf("get execution result err: %w", err)
	}
	log.Printf("GetExecutionResult status code: %d", statusCode)
	if statusCode != 200 {
		return nil, fmt.Errorf("request failed, status code: %d, body: %s", statusCode, string(body))
	}

	var execOutput volce_base.ExecutionAudioExtractResult
	err = json.Unmarshal(body, &execOutput)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response body err: %w", err)
	}
	return &execOutput, nil
}

func (v *VodVolce) SubmitAsrTask(vid, spaceName string) (string, error) {
	vodv2Session := volce_base.NewVodV2Session(v.accessKey, v.secretKey)

	startExecInput := &volce_base.StartExecutionInput{
		Input: &volce_base.InputForStartExecutionInput{
			Type: "Vid",
			Vid:  vid,
		},
		Operation: &volce_base.ConvertOperationForStartExecutionInput{
			Type: "Task",
			Task: &volce_base.Task{
				Type: "Asr",
				Asr: &volce_base.OperationTaskAsr{
					Type:            "speech",
					WithSpeakerInfo: true,
				},
			},
		},
		SpaceName: spaceName,
	}

	body, statusCode, err := vodv2Session.StartExecution(startExecInput)
	if err != nil {
		return "", fmt.Errorf("start execution err: %w", err)
	}
	log.Printf("StartExecution status code: %d", statusCode)
	if statusCode != 200 {
		return "", fmt.Errorf("request failed, status code: %d, body: %s", statusCode, string(body))
	}
	var startExecOutput volce_base.StartExecutionOutput
	err = json.Unmarshal(body, &startExecOutput)
	if err != nil {
		return "", fmt.Errorf("unmarshal response body err: %w", err)
	}
	return startExecOutput.Result.RunId, nil
}

func (v *VodVolce) QueryAsrTaskResult(runId string) (*volce_base.ExecutionAsrResult, error) {
	vodv2Session := volce_base.NewVodV2Session(v.accessKey, v.secretKey)

	body, statusCode, err := vodv2Session.GetExecutionResult(runId)
	if err != nil {
		return nil, fmt.Errorf("get execution result err: %w", err)
	}
	log.Printf("GetExecutionResult status code: %d", statusCode)
	if statusCode != 200 {
		return nil, fmt.Errorf("request failed, status code: %d, body: %s", statusCode, string(body))
	}

	var execOutput volce_base.ExecutionAsrResult
	err = json.Unmarshal(body, &execOutput)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response body err: %w", err)
	}
	return &execOutput, nil
}
