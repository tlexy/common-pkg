package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
)

const (
	ReqId    = "x-request-id"
	TenantId = "tenant-id"
)

type JsonOutputsForLogger struct {
	Time      string `json:"ts"`
	Caller    string `json:"caller"`
	RequestId string `json:"request_id"`
	Level     string `json:"level"`
	Msg       string `json:"msg"`
	Version   string `json:"version"` //版本号
}

func InitLog(file, logDir, version string, ctxKeys []interface{}) {
	defaultCtxKey := []interface{}{}
	// LoggingJsonHandler is a example handler for logging JSON format content.
	var LoggingJsonHandler glog.Handler = func(ctx context.Context, in *glog.HandlerInput) {
		jsonForLogger := JsonOutputsForLogger{
			Time:      in.TimeFormat,
			Level:     gstr.Trim(in.LevelFormat, "[]"),
			RequestId: in.TraceId,
			Caller:    gstr.Trim(fmt.Sprintf("%s%s", in.CallerPath, in.CallerFunc), ":"),
			Msg:       gstr.Trim(in.ValuesContent()),
			Version:   version,
		}
		if rId, ok := ctx.Value(ReqId).(string); ok {
			jsonForLogger.RequestId = rId
		}
		jsonBytes, err := json.Marshal(jsonForLogger)
		if err != nil {
			_, _ = os.Stderr.WriteString(err.Error())
			return
		}
		in.Buffer.Write(jsonBytes)
		in.Buffer.WriteString("\n")
		in.Next(ctx)
	}
	g.Log().SetHandlers(LoggingJsonHandler)
	cfg := g.Log().GetConfig()
	cfg.Level = glog.LEVEL_ALL
	cfg.Flags = cfg.Flags | glog.F_FILE_LONG | glog.F_CALLER_FN
	cfg.CtxKeys = append(cfg.CtxKeys, defaultCtxKey...)
	cfg.CtxKeys = append(cfg.CtxKeys, ctxKeys...)
	if logDir != "" {
		cfg.Path = logDir
		cfg.File = file
		cfg.StdoutPrint = false
	} else {
		cfg.StdoutPrint = true
	}
	g.Log().SetConfig(cfg)
}
