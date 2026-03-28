package voice_volce

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/google/uuid"
	"github.com/tlexy/common-pkg/util/xhttp"
)

// 核心接口
// 火山云语音识别

const (
	taskSuccess    = "20000000"
	taskInProgress = "20000001"
	taskQueued     = "20000002"

	XApiAppKey     = "X-Api-App-Key"
	XApiAccessKey  = "X-Api-Access-Key"
	XApiResourceId = "X-Api-Resource-Id"
	XApiRequestId  = "X-Api-Request-Id"
	XTtLogid       = "X-Tt-Logid"
	XVolcUser      = "X-Volc-User"
	XApiMessage    = "X-Api-Message"
	XApiStatusCode = "X-Api-Status-Code"
	submitUrl      = "https://openspeech.bytedance.com/api/v3/auc/bigmodel/submit"
	queryUrl       = "https://openspeech.bytedance.com/api/v3/auc/bigmodel/query"
)

type AiVoice struct {
	appKey     string
	accessKey  string
	httpClient *xhttp.XHttpClient
}

func NewAiVoice(appKey, accessKey string) *AiVoice {
	return &AiVoice{
		appKey:     appKey,
		accessKey:  accessKey,
		httpClient: xhttp.NewXHttpClient(),
	}
}

func (v *AiVoice) SubmitAsrTask(ctx context.Context, req *AsrRequest) (string, error) {
	if len(req.RequestId) < 3 {
		req.RequestId = uuid.New().String()
	}
	ari := &AsrRequestInternal{
		Audio: &AsrRequestInternalAudio{
			Url:      req.AudioUrl,
			Format:   string(req.AudioFormat),
			Language: req.Language,
		},
		Request: &AsrRequestInternalRequest{
			EnableSpeakerInfo:      true,
			ShowUtterances:         true,
			EnableLid:              true,
			EnableEmotionDetection: true,
			EnableGenderDetection:  true,
			VadSegment:             true,
			Callback:               req.CallbackUrl,
		},
	}
	reqHeader := map[string]string{
		XApiAppKey:       v.appKey,
		XApiAccessKey:    v.accessKey,
		XApiResourceId:   "volc.seedasr.auc",
		XApiRequestId:    req.RequestId,
		"X-Api-Sequence": "1",
	}

	respHeader := make(map[string]string)
	var response map[string]interface{} = make(map[string]interface{})
	err := retry.Do(func() error {
		return v.httpClient.PostJsonWithHeader(ctx, submitUrl, ari, &response, reqHeader, respHeader)
	},
		retry.Attempts(3),
		retry.Delay(3*time.Second),
		retry.MaxDelay(30*time.Second),
		retry.DelayType(retry.BackOffDelay),
	)
	log.Printf("VolcengineAsr submit response header: %+v response %+v\n", respHeader, response)
	if err != nil {
		log.Printf("VolcengineAsr task failed: %v\n", err)
		return "", err
	}

	statusCode := respHeader[XApiStatusCode]
	msg := respHeader[XApiMessage]
	logId := respHeader[XTtLogid]
	if statusCode != taskSuccess {
		log.Printf("VolcengineAsr task failed: req:%v code:%s logId:%s, msg:%s, err %v\n", req, statusCode, logId, msg, err)
		return "", err
	}

	return req.RequestId, nil
}

func (v *AiVoice) GetAsrResult(ctx context.Context, taskId string) (*AsrResultRes, error) {
	// https://www.volcengine.com/docs/6561/1354868?lang=zh
	reqHeader := map[string]string{
		XApiAppKey:       v.appKey,
		XApiAccessKey:    v.accessKey,
		XApiResourceId:   "volc.seedasr.auc",
		XApiRequestId:    taskId,
		"X-Api-Sequence": "1",
	}

	respHeader := make(map[string]string)
	var response AsrResultRes
	err := retry.Do(func() error {
		return v.httpClient.PostJsonWithHeader(ctx, queryUrl, make(map[string]interface{}), &response, reqHeader, respHeader)
	},
		retry.Attempts(3),
		retry.Delay(3*time.Second),
		retry.MaxDelay(30*time.Second),
		retry.DelayType(retry.BackOffDelay),
	)
	log.Printf("VolcengineAsr submit response header: %+v response %+v\n", respHeader, response)
	if err != nil {
		log.Printf("VolcengineAsr task failed: %v\n", err)
		return nil, err
	}

	statusCode := respHeader[XApiStatusCode]
	msg := respHeader[XApiMessage]
	logId := respHeader[XTtLogid]
	if statusCode != taskSuccess &&
		statusCode != taskInProgress &&
		statusCode != taskQueued {
		log.Printf("VolcengineAsr query task result failed: taskId:%v code:%s logId:%s, msg:%s, err %v\n",
			taskId, statusCode, logId, msg, err)
		return &AsrResultRes{
			TaskId:     taskId,
			StatusCode: AsrStatusFailed,
			Message:    fmt.Sprintf("query asr result failed, code:%s, logId:%s, msg:%s", statusCode, logId, msg),
		}, nil
	}
	if statusCode != taskSuccess {
		return &AsrResultRes{
			TaskId:     taskId,
			StatusCode: AsrStatusRunning,
			Message:    "task is still running",
		}, nil
	}

	return &response, nil
}
