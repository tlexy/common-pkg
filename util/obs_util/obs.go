package obs_utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tlexy/common-pkg/util/acl"
	"github.com/tlexy/common-pkg/util/common"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/google/uuid"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	hobs "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rds/v3/region"
)

const (
	ObsChunkSize int64 = 1024 * 3
)

func getRegionFromEndPoint(endPoint string) (string, error) {
	// EndPoint:   "obs.cn-north-4.myhuaweicloud.com",
	// 从endPoint中提取regionId
	pos := strings.Index(endPoint, ".")
	if pos == -1 {
		return "", fmt.Errorf("endPoint format error")
	}
	str := endPoint[pos+1:]
	pos2 := strings.Index(str, ".")
	if pos2 == -1 {
		return "", fmt.Errorf("endPoint format error: %s", str)
	}
	regionId := str[:pos2]
	return regionId, nil
}

func GetTemporaryCredentials(accessKey, secretKey, endPoint string, durationSeconds int32) (string, string, string, error) {
	regionId, err := getRegionFromEndPoint(endPoint)
	if err != nil {
		return "", "", "", err
	}
	return GetTemporaryCredentials2(accessKey, secretKey, regionId, durationSeconds)
}

func GetTemporaryCredentials2(accessKey, secretKey, regionId string, durationSeconds int32) (string, string, string, error) {
	// 创建认证信息
	globalAuth, err := basic.NewCredentialsBuilder().
		WithAk(accessKey).
		WithSk(secretKey).
		SafeBuild()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to NewCredentialsBuilder: %s", err)
	}

	region, err := region.SafeValueOf(regionId)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to region.SafeValueOf: %s", err)
	}
	// 初始化指定云服务的客户端 New{Service}Client ，以初始化 Global 级服务 IAM 的 IamClient 为例
	hcClient, err := iam.IamClientBuilder().
		WithRegion(region).
		WithCredential(globalAuth).
		WithHttpConfig(config.DefaultHttpConfig()).
		SafeBuild()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to IamClientBuilder: %s", err)
	}

	// 创建IAM客户端
	client := iam.NewIamClient(hcClient)

	// CreateTemporaryAccessKeyByToken
	id := uuid.New().String()

	identity := model.TokenAuthIdentity{
		Methods: []model.TokenAuthIdentityMethods{
			model.GetTokenAuthIdentityMethodsEnum().TOKEN,
		},
		Token: &model.IdentityToken{
			Id:              &id,
			DurationSeconds: &durationSeconds,
		},
	}
	body := model.CreateTemporaryAccessKeyByTokenRequestBody{
		Auth: &model.TokenAuth{
			Identity: &identity,
		},
	}
	request := model.CreateTemporaryAccessKeyByTokenRequest{
		Body: &body,
	}
	response, err := client.CreateTemporaryAccessKeyByToken(&request)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to CreateTemporaryAccessKeyByToken: %s", err)
	}
	credential := response.Credential

	return credential.Access, credential.Secret, credential.Securitytoken, nil
}

func RemoveObject(accessKey, secretKey, endPoint string, remoteBucket, objectName string) error {
	obsClient, err := hobs.New(accessKey, secretKey, endPoint, hobs.WithSignature(hobs.SignatureObs) /*, obs.WithSecurityToken(securityToken)*/)
	if err != nil {
		return err
	}
	input := &hobs.DeleteObjectInput{}
	input.Key = objectName
	input.Bucket = remoteBucket
	_, err = obsClient.DeleteObject(input)
	if err != nil {
		return err
	}
	return nil
}

func IsObjectExist(accessKey, secretKey, endPoint string, remoteBucket, objectName string) (bool, error) {
	obsClient, err := hobs.New(accessKey, secretKey, endPoint, hobs.WithSignature(hobs.SignatureObs) /*, obs.WithSecurityToken(securityToken)*/)
	if err != nil {
		return false, err
	}
	req := &hobs.GetObjectAclInput{}
	req.Key = objectName
	req.Bucket = remoteBucket
	resp, err := obsClient.GetObjectAcl(req)
	if err != nil {
		// g.Log().Warningf(ctx, "RemoteObjectIsExist, GetObjectAcl err: %s", err.Error())
		return false, err
	}
	if resp.StatusCode < 300 {
		return true, nil
	}
	return false, nil
}

// func RemoteObjectIsExist2(ctx context.Context, remoteUrl string) bool {
// 	return RemoteObjectIsExist(ctx, icfg.Cfg.Obs.ObsAccessKey, icfg.Cfg.Obs.ObsSecretKey, icfg.Cfg.Obs.ObsEndPoint, icfg.Cfg.Obs.ObsBucketName,
// 		remoteUrl)
// }
// func Download2(ctx context.Context, remoteUrl string, saveFullName string) error {
// 	return Download(ctx, icfg.Cfg.Obs.ObsAccessKey, icfg.Cfg.Obs.ObsSecretKey, icfg.Cfg.Obs.ObsEndPoint, icfg.Cfg.Obs.ObsBucketName,
// 		remoteUrl, saveFullName)
// }

// func ObsUpload2(objectUrl, srcFile, auxDir string) error {
// 	return ObsUpload(icfg.Cfg.Obs.ObsAccessKey, icfg.Cfg.Obs.ObsSecretKey, icfg.Cfg.Obs.ObsEndPoint, icfg.Cfg.Obs.ObsBucketName,
// 		objectUrl, srcFile, auxDir)
// }

// func SetPublic2(objectUrl string) error {
// 	return SetPublic(icfg.Cfg.Obs.ObsAccessKey, icfg.Cfg.Obs.ObsSecretKey, icfg.Cfg.Obs.ObsEndPoint, icfg.Cfg.Obs.ObsBucketName,
// 		objectUrl)
// }

// func SetImageMetadata2(objectUrl string) error {
// 	return SetImageMetadata(icfg.Cfg.Obs.ObsAccessKey, icfg.Cfg.Obs.ObsSecretKey, icfg.Cfg.Obs.ObsEndPoint, icfg.Cfg.Obs.ObsBucketName,
// 		objectUrl)
// }

func RemoteObjectIsExist(ctx context.Context, accessKey, secretKey, endPoint string, remoteBucket, remoteUrl string) bool {
	obsClient, err := hobs.New(accessKey, secretKey, endPoint, hobs.WithSignature(hobs.SignatureObs) /*, obs.WithSecurityToken(securityToken)*/)
	if err != nil {
		g.Log().Errorf(ctx, "RemoteObjectIsExist, err: %s", err.Error())
		return false
	}
	req := &hobs.GetObjectAclInput{}
	req.Key = remoteUrl
	req.Bucket = remoteBucket
	resp, err := obsClient.GetObjectAcl(req)
	if err != nil {
		g.Log().Errorf(ctx, "RemoteObjectIsExist, GetObjectAcl err: %s", err.Error())
		return false
	}
	if resp.StatusCode < 300 {
		return true
	}
	return false
}

// 从obs中下载文件，下载到哪里呢？当前文件夹下？下载完成后要不要删除呢？
// 保存到本地时，后缀名不能改变
func Download(ctx context.Context, accessKey, secretKey, endPoint string, remoteBucket, remoteUrl string, saveFullName string) error {
	obsClient, err := hobs.New(accessKey, secretKey, endPoint, hobs.WithSignature(hobs.SignatureObs))
	if err != nil {
		return err
	}
	input := &hobs.GetObjectInput{}
	// 指定存储桶名称
	input.Bucket = remoteBucket
	// 指定下载对象
	input.Key = remoteUrl

	// 流式下载对象
	output, err := obsClient.GetObject(input)
	if err != nil {
		return err
	}
	// output.Body 在使用完毕后必须关闭，否则会造成连接泄漏。
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			g.Log().Error(ctx, "Body.Close, err: %s", err.Error())
		}
	}(output.Body)

	// 新建本地文件
	// targetFile := common.StringFlag(remoteUrl)
	// fullPath := saveDir + "/" + targetFile + filepath.Ext(remoteUrl)
	file, err := os.Create(saveFullName)
	if err != nil || file == nil {
		if err == nil {
			return fmt.Errorf("download failed, create file is nil: %s", saveFullName)
		}
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			g.Log().Errorf(ctx, "file.Close err: %s", err.Error())
		}
	}(file)
	// 读取对象内容
	p := make([]byte, 4096)
	var readErr error
	var readCount int
	for {
		readCount, readErr = output.Body.Read(p)
		if readCount > 0 {
			writeCount, err := file.Write(p[:readCount])
			if err != nil {
				return err
			}
			if writeCount != readCount {
				return fmt.Errorf("Download failed, write file error, readCount: %d, wirteCount: %d", readCount, writeCount)
			}
		}
		if readErr != nil {
			break
		}
	}
	return nil
}

func DownloadBytes(ctx context.Context, accessKey, secretKey, endPoint string, remoteBucket, remoteUrl string) (io.ReadCloser, error) {
	obsClient, err := hobs.New(accessKey, secretKey, endPoint, hobs.WithSignature(hobs.SignatureObs))
	if err != nil {
		return nil, err
	}
	input := &hobs.GetObjectInput{}
	// 指定存储桶名称
	input.Bucket = remoteBucket
	// 指定下载对象
	input.Key = remoteUrl

	// 流式下载对象
	output, err := obsClient.GetObject(input)
	if err != nil {
		return nil, err
	}
	return output.Body, nil
	// output.Body 在使用完毕后必须关闭，否则会造成连接泄漏。
	// defer func(Body io.ReadCloser) {
	// 	err := Body.Close()
	// 	if err != nil {
	// 		g.Log().Error(ctx, "Body.Close, err: %s", err.Error())
	// 	}
	// }(output.Body)

	// 新建本地文件
	// targetFile := common.StringFlag(remoteUrl)
	// fullPath := saveDir + "/" + targetFile + filepath.Ext(remoteUrl)
	// file, err := os.Create(saveFullName)
	// if err != nil || file == nil {
	// 	if err == nil {
	// 		return fmt.Errorf("download failed, create file is nil: %s", saveFullName)
	// 	}
	// 	return err
	// }
	// defer func(file *os.File) {
	// 	err := file.Close()
	// 	if err != nil {
	// 		g.Log().Errorf(ctx, "file.Close err: %s", err.Error())
	// 	}
	// }(file)
	// 读取对象内容
	// p := make([]byte, 4096)
	// var readErr error
	// var readCount int
	// for {
	// 	readCount, readErr = output.Body.Read(p)
	// 	if readCount > 0 {
	// 		writeCount, err := file.Write(p[:readCount])
	// 		if err != nil {
	// 			return err
	// 		}
	// 		if writeCount != readCount {
	// 			return fmt.Errorf("Download failed, write file error, readCount: %d, wirteCount: %d", readCount, writeCount)
	// 		}
	// 	}
	// 	if readErr != nil {
	// 		break
	// 	}
	// }
	// return nil
}

func UploadBytes(accessKey, secretKey, endPoint string, remoteBucket, objectUrl string, srcFile []byte, aclType acl.StorageAclType) error {
	obsClient, err := hobs.New(accessKey, secretKey, endPoint, hobs.WithSignature(hobs.SignatureObs))
	if err != nil {
		return err
	}
	input := &hobs.PutObjectInput{}
	// 指定存储桶名称
	input.Bucket = remoteBucket
	// 指定上传对象，此处以 example/objectname 为例。
	input.Key = objectUrl
	// fd, _ := os.Open("localfile")
	input.Body = bytes.NewReader(srcFile)
	input.ACL = hobs.AclPrivate
	if aclType == acl.AclTypePublicRead {
		input.ACL = hobs.AclPublicRead
	}
	// 流式上传本地文件
	output, err := obsClient.PutObject(input)
	if err == nil {
		fmt.Printf("Put object(%s) under the bucket(%s) successful!\n", input.Key, input.Bucket)
		fmt.Printf("StorageClass:%s, ETag:%s\n",
			output.StorageClass, output.ETag)
		return nil
	}
	fmt.Printf("Put object(%s) under the bucket(%s) fail!\n", input.Key, input.Bucket)
	if obsError, ok := err.(obs.ObsError); ok {
		fmt.Println("An ObsError was found, which means your request sent to OBS was rejected with an error response.")
		// fmt.Println(obsError.Error())
		return obsError
	} else {
		fmt.Println("An Exception was found, which means the client encountered an internal problem when attempting to communicate with OBS, for example, the client was unable to access the network.")
		// fmt.Println(err)
		return err
	}
}

func ObsUpload(accessKey, secretKey, endPoint string, remoteBucket, objectUrl string, srcFile string, aclType acl.StorageAclType, auxDir string) error {
	// 1. 根据大小选择不同的上传函数
	// 2. 自动重试3次
	fi, err := os.Stat(srcFile)
	if err != nil {
		return fmt.Errorf("上传时找不到文件， file: %s", srcFile)
	}
	retriesLimit := 2
	retriesCount := 0
	errMsg := ""
	for {
		if retriesCount > retriesLimit {
			return fmt.Errorf("上传失败, errMsg: %s, retriesCount: %v", errMsg, retriesCount)
		}
		if fi.Size() < 1024*1024*ObsChunkSize {
			err = uploadSmallFile(accessKey, secretKey, endPoint, remoteBucket, objectUrl, srcFile, aclType)
		} else {
			err = uploadBigFile(accessKey, secretKey, endPoint, remoteBucket, objectUrl, srcFile, auxDir, aclType)
		}
		if err == nil {
			return nil
		}
		retriesCount++
	}
}

func uploadSmallFile(accessKey, secretKey, endPoint string, remoteBucket, objectUrl string, srcFile string, aclType acl.StorageAclType) error {
	obsClient, err := hobs.New(accessKey, secretKey, endPoint, hobs.WithSignature(hobs.SignatureObs))
	if err != nil {
		return err
	}

	input := &hobs.PutFileInput{}
	// 指定存储桶名称
	input.Bucket = remoteBucket
	// 指定上传对象，此处以 example/objectname 为例。
	input.Key = objectUrl
	// 指定本地文件，此处以localfile为例
	input.SourceFile = srcFile
	input.ACL = hobs.AclPrivate
	if aclType == acl.AclTypePublicRead {
		input.ACL = hobs.AclPublicRead
	}
	// 文件上传
	output, err := obsClient.PutFile(input)
	if err != nil {
		// log.Error("ObsUpload failed", zap.Error(err), zap.String("object", input.Key), zap.String("bucket", input.Bucket))
		return fmt.Errorf("Put file(%s) under the bucket(%s) error: %s!\n", input.Key, input.Bucket, err.Error())
	}
	if output == nil {
		// log.Error("ObsUpload failed, output is nil", zap.String("object", input.Key), zap.String("bucket", input.Bucket))
		return fmt.Errorf("Put file(%s) under the bucket(%s) failed, output is nil!\n", input.Key, input.Bucket)
	}
	return nil
}

func uploadBigFile(accessKey, secretKey, endPoint string, remoteBucket, objectUrl string, srcFile, auxDir string, aclType acl.StorageAclType) error {
	// 上传大文件@https://support.huaweicloud.com/intl/zh-cn/sdk-go-devg-obs/obs_33_0521.html

	err := common.EnsureOutputDirectory(auxDir)
	if err != nil {
		return fmt.Errorf("上传文件，目录创建错误, dir: %s, err: %s", auxDir, err.Error())
	}
	err = os.RemoveAll(auxDir)
	if err != nil {
		return fmt.Errorf("删除目录出错, dir: %s, err: %s", auxDir, err.Error())
	}
	err = common.EnsureOutputDirectory(auxDir)
	if err != nil {
		return fmt.Errorf("创建目录出错, dir: %s, err: %s", auxDir, err.Error())
	}
	// 对输入的文件进行分块
	cmd := exec.Command("split",
		"-b", fmt.Sprintf("%vM", ObsChunkSize),
		"--additional-suffix=.mp4",
		srcFile,
		"-d",
		fmt.Sprintf("%s/obs_chunk_", auxDir))
	fmt.Println("Running command:", strings.Join(cmd.Args, " "))
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("拆分大文件失败, srcFile: %s, err: %s, output: %s", srcFile, err.Error(), stdoutStderr)
	}
	// 获取目录下的所有文件并排序
	files, err := os.ReadDir(auxDir)
	if err != nil {
		return fmt.Errorf("读取目录信息失败, err: %s", err.Error())
	}
	var chunks []string
	for _, val := range files {
		if val.IsDir() {
			continue
		}
		if filepath.Ext(val.Name()) == ".mp4" {
			chunks = append(chunks, val.Name())
		}
	}
	sort.Strings(chunks)
	uploadChunks := make([]string, 0, len(chunks))
	for _, val := range chunks {
		uploadChunks = append(uploadChunks, fmt.Sprintf("%s/%s", auxDir, val))
	}
	if len(uploadChunks) < 2 {
		return fmt.Errorf("分片数量太少， size: %v", len(uploadChunks))
	}

	obsClient, err := hobs.New(accessKey, secretKey, endPoint, hobs.WithSignature(hobs.SignatureObs))
	if err != nil {
		return err
	}

	// 初始化分段上传任务
	inputInit := &hobs.InitiateMultipartUploadInput{}
	inputInit.Bucket = remoteBucket
	inputInit.Key = objectUrl
	inputInit.ACL = hobs.AclPrivate
	if aclType == acl.AclTypePublicRead {
		inputInit.ACL = hobs.AclPublicRead
	}
	outputInit, err := obsClient.InitiateMultipartUpload(inputInit)
	if err != nil {
		if obsError, ok := err.(hobs.ObsError); ok {
			return fmt.Errorf("初始化分段失败, code: %v, msg: %s", obsError.Code, obsError.Message)
		} else {
			return err
		}
	}
	// 分片上传
	uploadId := outputInit.UploadId
	etags := make([]string, 0, len(uploadChunks))
	for idx, val := range uploadChunks {
		inputUploadPart := &hobs.UploadPartInput{}
		inputUploadPart.Bucket = remoteBucket
		inputUploadPart.Key = objectUrl
		inputUploadPart.UploadId = uploadId
		inputUploadPart.PartNumber = idx + 1
		inputUploadPart.SourceFile = val
		outputUploadPart, err := obsClient.UploadPart(inputUploadPart)
		if err != nil {
			if obsError, ok := err.(hobs.ObsError); ok {
				return fmt.Errorf("分段上传失败, idx: %v, fn: %s, code: %v, msg: %s", idx, val, obsError.Code, obsError.Message)
			} else {
				return err
			}
		} else {
			etags = append(etags, outputUploadPart.ETag)
		}
	}
	// 合并段
	Parts := make([]hobs.Part, 0, len(etags))
	if len(etags) != len(uploadChunks) {
		return fmt.Errorf("etags大小不一致，%v:%v", len(etags), len(uploadChunks))
	}
	for idx, val := range etags {
		Parts = append(Parts, hobs.Part{PartNumber: idx + 1, ETag: val})
	}
	inputCompleteMultipart := &hobs.CompleteMultipartUploadInput{}
	inputCompleteMultipart.Bucket = remoteBucket
	inputCompleteMultipart.Key = objectUrl
	inputCompleteMultipart.UploadId = uploadId
	inputCompleteMultipart.Parts = Parts
	_, err = obsClient.CompleteMultipartUpload(inputCompleteMultipart)
	if err != nil {
		if obsError, ok := err.(hobs.ObsError); ok {
			return fmt.Errorf("合并分段失败, err: %s, code: %v, msg: %s", err.Error(), obsError.Code, obsError.Message)
		} else {
			return err
		}
	}
	err = os.RemoveAll(auxDir)
	if err != nil {
		fmt.Printf("上传大文件后删除目录出错, dir: %s, err: %s", auxDir, err.Error())
	}
	return nil
}

func SetPublic(accessKey, secretKey, endPoint string, remoteBucket, objectUrl string) error {
	obsClient, err := hobs.New(accessKey, secretKey, endPoint, hobs.WithSignature(hobs.SignatureObs))
	if err != nil {
		// log.Error("SetPublic", zap.Error(err))
		return err
	}

	acl := &hobs.SetObjectAclInput{}
	acl.Key = objectUrl
	acl.Bucket = remoteBucket
	acl.ACL = hobs.AclPublicRead

	res, err := obsClient.SetObjectAcl(acl)
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("res is nil, objectUrl: %s", objectUrl)
	}
	if res.StatusCode < 300 {
		return nil
	} else {
		return fmt.Errorf("res_code is %d, objectUrl: %s", res.StatusCode, objectUrl)
	}
}

func SetImageMetadata(accessKey, secretKey, endPoint string, remoteBucket, objectUrl string) error {
	obsClient, err := hobs.New(accessKey, secretKey, endPoint, hobs.WithSignature(hobs.SignatureObs))
	if err != nil {
		// log.Error("setImageMetadata", zap.Error(err))
		return err
	}

	input := &hobs.SetObjectMetadataInput{}
	// 指定存储桶名称
	input.Bucket = remoteBucket
	// 指定对象，此处以 example/objectname 为例。
	input.Key = objectUrl
	// 指定对象MIME类型，这里以image/jpeg为例
	input.ContentType = "image/jpeg"
	input.ContentDisposition = "inline"
	input.StorageClass = hobs.StorageClassStandard

	_, err = obsClient.SetObjectMetadata(input)
	if err == nil {
		// fmt.Printf("Set Object(%s)'s metadata successful with bucket(%s)!\n", input.Key, input.Bucket)
		// fmt.Printf("RequestId:%s\n", output.RequestId)
		return nil
	}
	// log.Error("setImageMetadata error", zap.Error(err))
	return fmt.Errorf("Set Object(%s)'s metadata fail with bucket(%s)!\n", input.Key, input.Bucket)
	//if obsError, ok := err.(obs.ObsError); ok {
	//	fmt.Println("An ObsError was found, which means your request sent to OBS was rejected with an error response.")
	//	fmt.Println(obsError.Error())
	//} else {
	//	fmt.Println("An Exception was found, which means the client encountered an internal problem when attempting to communicate with OBS, for example, the client was unable to access the network.")
	//	fmt.Println(err)
	//}
}
