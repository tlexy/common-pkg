package tos_util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	cacl "github.com/tlexy/common-pkg/util/acl"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
	"github.com/volcengine/volc-sdk-golang/service/sts"
)

func GetTemporaryCredentials(accessKey, secretKey, endPoint, bucketName, tosTrn string, sec int64) (string, string, string, error) {
	//参考： https://www.volcengine.com/docs/6349/127695?lang=zh
	sts.DefaultInstance.Client.SetAccessKey(accessKey)
	sts.DefaultInstance.Client.SetSecretKey(secretKey)

	list, status, err := sts.DefaultInstance.AssumeRole(&sts.AssumeRoleRequest{
		DurationSeconds: int(sec),
		RoleTrn:         tosTrn,
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

func checkErr(err error) {
	if err != nil {
		if serverErr, ok := err.(*tos.TosServerError); ok {
			fmt.Println("Error:", serverErr.Error())
			fmt.Println("Request ID:", serverErr.RequestID)
			fmt.Println("Response Status Code:", serverErr.StatusCode)
			fmt.Println("Response Header:", serverErr.Header)
			fmt.Println("Response Err Code:", serverErr.Code)
			fmt.Println("Response Err Msg:", serverErr.Message)
		} else if clientErr, ok := err.(*tos.TosClientError); ok {
			fmt.Println("Error:", clientErr.Error())
			fmt.Println("Client Cause Err:", clientErr.Cause.Error())
		} else {
			fmt.Println("Error:", err)
		}
		//panic(err)
	}
}

func getRegionByEndpoint(endpoint string) string {
	// 解析endpoint获取region
	// endpoint格式：tos-{region}.ivolces.com
	pos := strings.Index(endpoint, "tos-")
	if pos == -1 {
		return ""
	}
	endPos := strings.Index(endpoint[pos:], ".")
	if endPos == -1 {
		return ""
	}
	return endpoint[pos+len("tos-") : pos+endPos]
}

func Upload(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, localPath string, objectName string, aclType cacl.StorageAclType) error {
	_, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("上传时找不到文件， file: %s", localPath)
	}
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}
	// 初始化客户端
	client, err := tos.NewClientV2(endPoint, tos.WithRegion(region), tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)))
	if err != nil {
		checkErr(err)
		return fmt.Errorf("创建tos客户端失败, err: %w", err)
	}
	// 读取本地文件数据
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败, file: %s, err: %w", localPath, err)
	}
	defer f.Close()
	acl := enum.ACLPrivate
	if aclType == cacl.AclTypePublicRead {
		acl = enum.ACLPublicRead
	}
	output, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucketName,
			Key:    objectName,
			ACL:    acl,
		},
		Content: f,
	})
	// 检查错误
	if err != nil {
		checkErr(err)
		return fmt.Errorf("上传文件到tos失败, err: %w", err)
	}
	fmt.Println("PutObjectV2 Request ID:", output.RequestID)
	return nil
}

func UploadBytesByReader(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, reader io.Reader, objectName string, aclType cacl.StorageAclType) error {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}
	// 初始化客户端
	client, err := tos.NewClientV2(endPoint, tos.WithRegion(region), tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)))
	if err != nil {
		checkErr(err)
		return fmt.Errorf("创建tos客户端失败, err: %w", err)
	}
	output, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucketName,
			Key:    objectName,
		},
		Content: reader,
	})
	// 检查错误
	if err != nil {
		checkErr(err)
		return fmt.Errorf("上传文件到tos失败, err: %w", err)
	}
	fmt.Println("PutObjectV2 Request ID:", output.RequestID)
	return nil
}

func Download(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, objectName string, localPath string) error {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}

	// 初始化客户端
	client, err := tos.NewClientV2(endPoint, tos.WithRegion(region), tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)))
	checkErr(err)

	// 下载文件到指定的路径，示例中下载文件到 example_dir/example.txt
	getObjectToFileOutput, err := client.GetObjectToFile(ctx, &tos.GetObjectToFileInput{
		GetObjectV2Input: tos.GetObjectV2Input{
			Bucket: bucketName,
			Key:    objectName,
		},
		FilePath: localPath,
	})
	checkErr(err)
	fmt.Println("GetObjectToFile Request ID:", getObjectToFileOutput.RequestID)
	fmt.Println("GetObjectToFile File Size:", getObjectToFileOutput.ContentLength)
	return nil
}

func DownloadBytes(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, objectName string) (io.ReadCloser, error) {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return nil, fmt.Errorf("endpoint is not valid")
	}

	// 初始化客户端
	client, err := tos.NewClientV2(endPoint, tos.WithRegion(region), tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)))
	if err != nil {
		checkErr(err)
		return nil, fmt.Errorf("创建tos客户端失败, err: %w", err)
	}
	// 下载数据到内存
	getOutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: bucketName,
		Key:    objectName,
		// 获取当前下载进度
		DataTransferListener: nil,
		// 配置客户端限制
		RateLimiter:             nil,
		ResponseContentEncoding: "deflate",
	})
	if err != nil {
		checkErr(err)
		return nil, fmt.Errorf("下载文件失败, err: %w", err)
	}
	fmt.Println("GetObjectV2 Request ID:", getOutput.RequestID)
	// 下载数据大小
	fmt.Println("GetObjectV2 Response ContentLength", getOutput.ContentLength)
	if getOutput.Content != nil {
		return getOutput.Content, nil
	}
	return nil, fmt.Errorf("下载文件失败, Content is nil")
}
