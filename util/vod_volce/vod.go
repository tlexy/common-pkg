package vod_volce

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/volcengine/volc-sdk-golang/base"
	"github.com/volcengine/volc-sdk-golang/service/vod"
	"github.com/volcengine/volc-sdk-golang/service/vod/models/business"
	"github.com/volcengine/volc-sdk-golang/service/vod/models/request"
	"github.com/volcengine/volc-sdk-golang/service/vod/upload/functions"
)

type VodVolce struct {
	accessKey string
	secretKey string
	spaceName string
}

func NewVodVolce(accessKey, secretKey, spaceName string) *VodVolce {
	return &VodVolce{
		accessKey: accessKey,
		secretKey: secretKey,
		spaceName: spaceName,
	}
}

func (v *VodVolce) UploadMedia(localFilename string) (string, error) {
	return v.UploadMediaWithName(localFilename, "")
}

func (v *VodVolce) UploadMediaWithName(localFilename, objectName string) (string, error) {
	instance := vod.NewInstance()
	instance.SetCredential(base.Credentials{
		AccessKeyID:     v.accessKey,
		SecretAccessKey: v.secretKey,
	})

	spaceName := v.spaceName

	title := filepath.Base(localFilename)
	format := filepath.Ext(localFilename)
	optionFunc := functions.AddOptionInfoFunc(business.VodUploadFunctionInput{
		Title:            title,          // 视频的标题
		Tags:             "upload video", // 视频的标签
		Description:      "upload video", // 视频的描述信息
		Format:           format,         // 音视频格式
		ClassificationId: 0,              // 分类 Id，上传时可以指定分类，非必须字段
		IsHlsIndexOnly:   false,          //该字段传true表示视频仅关联hls文件，删除时不会删除ts文件
	})

	vodFunctions := []business.VodUploadFunction{optionFunc}
	fbts, _ := json.Marshal(vodFunctions)

	vodUploadMediaRequset := &request.VodUploadMediaRequest{
		SpaceName:        spaceName,     // 空间名称
		FilePath:         localFilename, // 本地文件路径
		CallbackArgs:     "",            // 透传信息，业务希望透传的字段可以写入，返回和回调中会返回此字段，非必须字段
		Functions:        string(fbts),  // 函数功能，具体可以参考火山引擎点播文档 开发者API-媒资上传-确认上传的 Functions 部分，可选功能字段
		FileName:         localFilename, // 设置文件名，无格式长度限制，用户可自定义,目前文件名不支持空格、+ 字符,如果要使用此字段，请联系技术支持配置白名单，非必须字段
		FileExtension:    "",            // 设置文件后缀，以 . 开头，不超过8位
		VodUploadSource:  "upload",      // 设置上传来源，值为枚举值
		ParallelNum:      2,             // 开启2协程进行分片上传，不配置时默认单协程，可根据机器 cpu 内存配置进行协程数设置
		UploadHostPrefer: "",            // 设置上传域名偏好
	}

	resp, _, err := instance.UploadMediaWithCallback(vodUploadMediaRequset)
	if err != nil {
		fmt.Printf("UploadMediaWithCallback error %v", err)
		return "", err
	}

	fmt.Println()
	fmt.Println(resp.GetResponseMetadata().GetRequestId())
	fmt.Println(resp.GetResult().GetData().GetVid())
	fmt.Println(resp.GetResult().GetData().GetSourceInfo().GetFileName())

	return resp.GetResult().GetData().GetVid(), nil
}
