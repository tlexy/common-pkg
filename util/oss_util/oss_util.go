package oss_util

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
	"github.com/gogf/gf/v2/frame/g"

	cacheClient "github.com/alibabacloud-go/cdn-20180510/v7/client"
	cacheApi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/tlexy/common-pkg/util/acl"
	"github.com/tlexy/common-pkg/util/common"
)

// https://help.aliyun.com/zh/oss/user-guide/regions-and-endpoints?spm=a2c4g.11186623.0.0.497f7368lOwdrL#concept-zt4-cvy-5db
var ossRegions = []string{
	"cn-shanghai",
	"cn-nanjing",
	"cn-fuzhou",
	"cn-wuhan-lr",
	"cn-huhehaote",
	"cn-qingdao",
	"cn-beijing",
	"cn-zhangjiakou",
	"cn-heyuan",
	"cn-shenzhen",
	"cn-wulanchabu",
	"cn-guangzhou",
	"cn-chengdu",
	"cn-hongkong",
	"ap-northeast-1",
	"ap-northeast-2",
	"ap-southeast-1",
	"ap-southeast-3",
	"ap-southeast-5",
	"ap-southeast-6",
	"ap-southeast-7",
	"eu-central-1",
	"eu-west-1",
	"us-west-1",
	"us-east-1",
	"me-east-1",
}

const (
	OssChunkSize int64 = 1024 * 3
)

var (
	clientCache          = make(map[string]*cacheClient.Client)
	clientCacheMutex     sync.Mutex
	CdnCacheAutoRetry    = false
	CdnCacheMaxAttempts  = 3
	CdnCacheKeepAlive    = true
	CdnCacheMaxIdleConns = 20
)

type headerConfig struct {
	ContentDisposition string
	Expires            int32 // 过期秒数
	PushCache          bool  // 是否开启CDN缓存
}

type ClientOption func(cfg *headerConfig)

func WithContentDisposition(contentDisposition string) ClientOption {
	return func(cfg *headerConfig) {
		cfg.ContentDisposition = contentDisposition
	}
}

func WithExpires(expires int32) ClientOption {
	return func(cfg *headerConfig) {
		cfg.Expires = expires
	}
}

func WithPushCache(pushCache bool) ClientOption {
	return func(cfg *headerConfig) {
		cfg.PushCache = pushCache
	}
}

type pushObjectCacheConfig struct {
	AutoRetry    *bool
	MaxAttempts  *int
	MaxIdleConns *int
}

type PushObjectCacheOption func(cfg *pushObjectCacheConfig)

func WithAutoRetry(autoRetry bool) PushObjectCacheOption {
	return func(cfg *pushObjectCacheConfig) {
		cfg.AutoRetry = &autoRetry
	}
}

func WithMaxAttempts(maxAttempts int) PushObjectCacheOption {
	return func(cfg *pushObjectCacheConfig) {
		cfg.MaxAttempts = &maxAttempts
	}
}

func WithMaxIdleConns(maxIdleConns int) PushObjectCacheOption {
	return func(cfg *pushObjectCacheConfig) {
		cfg.MaxIdleConns = &maxIdleConns
	}
}

func getRegionByEndpoint(endpoint string) string {
	for _, region := range ossRegions {
		if strings.Contains(endpoint, region) {
			return region
		}
	}
	return ""
}

// GetOSSCDNClient 返回CDN客户端，如果已存在则返回缓存的实例，否则创建新实例
func GetOSSCDNClient(ctx context.Context, accessKey, secretKey string) (*cacheClient.Client, error) {
	// Create a cache key from accessKey and secretKey
	cacheKey := accessKey + ":" + secretKey

	// Check if we already have a client for this key combination
	clientCacheMutex.Lock()
	if client, exists := clientCache[cacheKey]; exists {
		clientCacheMutex.Unlock()
		return client, nil
	}
	clientCacheMutex.Unlock()

	// Create a new client if not found in cache
	config := new(credentials.Config).
		SetType("access_key").
		SetAccessKeyId(accessKey).
		SetAccessKeySecret(secretKey)

	akCredential, err := credentials.NewCredential(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %v", err)
	}

	cacheConfig := &cacheApi.Config{
		Credential: akCredential,
		Endpoint:   tea.String("cdn.aliyuncs.com"),
	}

	cli, clientErr := cacheClient.NewClient(cacheConfig)
	if clientErr != nil {
		g.Log().Errorf(ctx, "failed to create CDN client: %v", clientErr)
		return nil, fmt.Errorf("failed to create CDN client: %v", clientErr)
	}

	// Store the new client in cache
	clientCacheMutex.Lock()
	clientCache[cacheKey] = cli
	clientCacheMutex.Unlock()

	return cli, nil
}

func GetTemporaryCredentials(accessKey, secretKey, endPoint, arn string, durationSeconds int32) (string, string, string, error) {
	regionId := getRegionByEndpoint(endPoint)
	stsClient, err := sts.NewClientWithAccessKey(regionId, accessKey, secretKey)
	if err != nil {
		return "", "", "", err
	}
	// 创建获取临时凭证的请求
	request := sts.CreateAssumeRoleRequest()
	request.Scheme = "https"
	request.RoleArn = arn                                                                                    // callerRes.Arn //"acs:ram::account_id007:role/role_temp_01"
	request.RoleSessionName = time.Now().Format("20060615") + "_" + strconv.FormatInt(time.Now().Unix(), 10) // callerRes.RoleId
	request.DurationSeconds = requests.NewInteger(int(durationSeconds))                                      // 设置临时凭证的有效期，单位为秒

	// 发送请求并获取临时凭证
	response, err := stsClient.AssumeRole(request)
	if err != nil {
		return "", "", "", err
	}

	// 打印临时凭证
	// fmt.Println("AccessKeyId:", response.Credentials.AccessKeyId)
	// fmt.Println("AccessKeySecret:", response.Credentials.AccessKeySecret)
	// fmt.Println("SecurityToken:", response.Credentials.SecurityToken)
	return response.Credentials.AccessKeyId, response.Credentials.AccessKeySecret, response.Credentials.SecurityToken, nil
}

func RemoveObject(accessKey, secretKey, endPoint, bucketName string, objectName string) error {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)
	// 创建删除对象的请求
	request := &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}
	// 执行删除对象的请求
	_, err := client.DeleteObject(context.TODO(), request)
	if err != nil {
		return err
	}
	return nil
}

func IsObjectExist(accessKey, secretKey, endPoint, bucketName string, objectName string) (bool, error) {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return false, fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)
	// 创建OSS客户端
	client := oss.NewClient(cfg)

	return client.IsObjectExist(context.TODO(), bucketName, objectName)
}

func GetObjectSize(accessKey, secretKey, endPoint, bucketName string, objectName string) (int64, error) {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return -1, fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)
	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建获取对象元数据的请求
	request := &oss.GetObjectMetaRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}
	// 执行获取对象元数据的请求
	resp, err := client.GetObjectMeta(context.TODO(), request)
	if err != nil {
		return -1, err
	}
	return resp.ContentLength, nil
}

func SetSymlink(accessKey, secretKey, endPoint, bucketName, objectName, symlinkName string) (string, error) {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return "", fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)
	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建设置软链接的请求
	putRequest := &oss.PutSymlinkRequest{
		Bucket: oss.Ptr(bucketName),  // 存储空间名称
		Key:    oss.Ptr(symlinkName), // 填写软链接名称
		Target: oss.Ptr(objectName),  // 填写软链接的目标文件名称
	}

	// 执行设置软链接的请求
	putResult, err := client.PutSymlink(context.TODO(), putRequest)
	if err != nil {
		log.Fatalf("failed to put symlink %v", err)
	}

	// 打印设置软链接的结果
	log.Printf("put symlink result:%#v\n", putResult)
	return "", nil
}

func Download(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, objectName string, localPath string) error {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)
	// 创建OSS客户端
	client := oss.NewClient(cfg)
	// 创建下载对象的请求
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}
	// 执行下载对象的请求
	resp, err := client.GetObject(context.TODO(), request)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("下载失败, result is nil")
	}
	defer resp.Body.Close()
	// 读取下载对象的数据
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// // 将下载的数据写入本地文件
	// err = os.WriteFile(localPath, body, 0644)
	// if err != nil {
	// 	return err
	// }
	fd, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o660)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = io.Copy(fd, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func DownloadBytes(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, objectName string) (io.ReadCloser, error) {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return nil, fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)
	// 创建OSS客户端
	client := oss.NewClient(cfg)
	// 创建下载对象的请求
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}
	// 执行下载对象的请求
	resp, err := client.GetObject(context.TODO(), request)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, fmt.Errorf("下载失败, result is nil")
	}
	return resp.Body, nil
}

func UploadBytesByReader(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, reader io.Reader, objectName string, aclType acl.StorageAclType, options ...ClientOption) error {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)
	// 创建OSS客户端
	client := oss.NewClient(cfg)

	aclTypeStr := oss.ObjectACLDefault
	if aclType == acl.AclTypePublicRead {
		aclTypeStr = oss.ObjectACLPublicRead
	}

	// 创建上传对象的请求
	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
		Body:   reader,
		Acl:    aclTypeStr,
	}
	headerConfig := &headerConfig{}
	for _, option := range options {
		option(headerConfig)
	}
	if headerConfig.ContentDisposition != "" {
		request.ContentDisposition = &headerConfig.ContentDisposition
	}
	// 执行上传对象的请求
	result, err := client.PutObject(context.TODO(), request)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("上传失败, result is nil")
	}
	return nil
}

func UploadByReadCloser(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, reader io.ReadCloser, objectName string, aclType acl.StorageAclType) error {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)
	// 创建OSS客户端
	client := oss.NewClient(cfg)

	defer reader.Close()

	aclTypeStr := oss.ObjectACLDefault
	if aclType == acl.AclTypePublicRead {
		aclTypeStr = oss.ObjectACLPublicRead
	}
	// 创建上传对象的请求
	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
		Body:   reader,
		Acl:    aclTypeStr,
	}
	// 执行上传对象的请求
	result, err := client.PutObject(context.TODO(), request)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("上传失败, result is nil")
	}
	return nil
}

func Upload(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, localPath string, objectName string, aclType acl.StorageAclType, options ...ClientOption) error {
	fi, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("上传时找不到文件， file: %s", localPath)
	}

	if fi.Size() < 1024*1024*OssChunkSize {
		// return uploadSmallFile(ctx, accessKey, secretKey, endPoint, bucketName, localPath, objectName, aclType)
		// 小于1M，直接上传
		return uploadSmallFile(ctx, accessKey, secretKey, endPoint, bucketName, localPath, objectName, aclType, options...)
	} else {
		// 大于1M，分块上传
		return uploadLargeFile(ctx, accessKey, secretKey, endPoint, bucketName, localPath, objectName, aclType)
	}
}

func uploadLargeFile(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, localPath string, objectName string, aclType acl.StorageAclType) error {
	currDir := filepath.Dir(localPath)
	auxDir := currDir + "/upload_chunks"
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

	defer os.RemoveAll(auxDir)

	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)

	// 对输入的文件进行分块
	cmd := exec.Command("split",
		"-b", fmt.Sprintf("%vM", OssChunkSize),
		"--additional-suffix=.mp4",
		localPath,
		"-d",
		fmt.Sprintf("%s/oss_chunk_", auxDir))
	fmt.Println("Running command:", strings.Join(cmd.Args, " "))
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("拆分大文件失败, srcFile: %s, err: %s, output: %s", localPath, err.Error(), stdoutStderr)
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

	// 创建OSS客户端
	client := oss.NewClient(cfg)
	// 创建分块上传的请求
	request := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}
	// 初始化分块上传
	initResult, err := client.InitiateMultipartUpload(context.TODO(), request)
	if err != nil {
		return err
	}
	if initResult == nil {
		return fmt.Errorf("初始化分块上传失败, result is nil")
	}
	uploadId := *initResult.UploadId

	var parts []oss.UploadPart
	// 启动多个goroutine进行分片上传
	for idx, val := range uploadChunks {
		// 读取文件
		file, err := os.Open(val)
		if err != nil {
			log.Fatalf("failed to open local file %v", err)
		}
		defer file.Close()
		// 创建分片上传请求
		partRequest := &oss.UploadPartRequest{
			Bucket:     oss.Ptr(bucketName), // 目标存储空间名称
			Key:        oss.Ptr(objectName), // 目标对象名称
			PartNumber: int32(idx + 1),      // 分片编号
			UploadId:   oss.Ptr(uploadId),   // 上传ID
			Body:       file,                // 分片文件路径
		}

		// 发送分片上传请求
		partResult, err := client.UploadPart(context.TODO(), partRequest)
		if err != nil {
			fmt.Errorf("failed to upload part %d: %v", partRequest.PartNumber, err)
		}

		// 记录分片上传结果
		part := oss.UploadPart{
			PartNumber: partRequest.PartNumber,
			ETag:       partResult.ETag,
		}
		parts = append(parts, part)
	}

	aclTypeStr := oss.ObjectACLDefault
	if aclType == acl.AclTypePublicRead {
		aclTypeStr = oss.ObjectACLPublicRead
	}
	if aclType == acl.AclTypePrivate {
		aclTypeStr = oss.ObjectACLPrivate
	}

	// 完成分片上传请求
	crequest := &oss.CompleteMultipartUploadRequest{
		Bucket:   oss.Ptr(bucketName),
		Key:      oss.Ptr(objectName),
		UploadId: oss.Ptr(uploadId),
		CompleteMultipartUpload: &oss.CompleteMultipartUpload{
			Parts: parts,
		},
		Acl: aclTypeStr,
	}
	cresult, err := client.CompleteMultipartUpload(context.TODO(), crequest)
	if err != nil {
		log.Fatalf("failed to complete multipart upload %v", err)
	}
	if cresult == nil {
		return fmt.Errorf("完成分片上传失败, result is nil")
	}
	return nil
}

// func uploadSmallFile(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, localPath string, objectName string, aclType acl.StorageAclType) error {
func uploadSmallFile(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, localPath string, objectName string, aclType acl.StorageAclType, options ...ClientOption) error {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	reader, err := os.Open(localPath)
	if err != nil {
		return err
	}

	defer reader.Close()

	aclTypeStr := oss.ObjectACLDefault
	if aclType == acl.AclTypePublicRead {
		aclTypeStr = oss.ObjectACLPublicRead
	}
	if aclType == acl.AclTypePrivate {
		aclTypeStr = oss.ObjectACLPrivate
	}
	// 创建上传对象的请求
	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
		Body:   reader,
		Acl:    aclTypeStr,
	}
	headerConfig := &headerConfig{}
	for _, option := range options {
		option(headerConfig)
	}
	if headerConfig.ContentDisposition != "" {
		request.ContentDisposition = &headerConfig.ContentDisposition
	}
	// 执行上传对象的请求
	result, err := client.PutObject(context.TODO(), request)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("上传失败, result is nil")
	}
	return nil
}

func SetPublic(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, objectName string) error {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建上传对象的请求
	putRequest := &oss.PutObjectAclRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
		Acl:    oss.ObjectACLPublicRead,
	}
	putResult, err := client.PutObjectAcl(context.TODO(), putRequest)
	if err != nil {
		// log.Fatalf("failed to put object acl %v", err)
		return err
	}
	if putResult == nil {
		return fmt.Errorf("设置失败, result is nil")
	}
	return nil
}

func SetPrivate(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, objectName string) error {

	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return fmt.Errorf("endpoint is not valid")
	}
	//通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建上传对象的请求
	putRequest := &oss.PutObjectAclRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
		Acl:    oss.ObjectACLPrivate,
	}
	putResult, err := client.PutObjectAcl(context.TODO(), putRequest)
	if err != nil {
		//log.Fatalf("failed to put object acl %v", err)
		return err
	}
	if putResult == nil {
		return fmt.Errorf("设置失败, result is nil")
	}
	return nil
}

func Presign(ctx context.Context, accessKey, secretKey, endPoint, bucketName string, objectName string, options ...ClientOption) (string, error) {
	region := getRegionByEndpoint(endPoint)
	if region == "" {
		return "", fmt.Errorf("endpoint is not valid")
	}
	// 通过accesskey以及access secret创建一个OSSClient实例
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(NewXCredentialsProvider(accessKey, secretKey)).
		WithEndpoint(endPoint).WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)
	headerConfig := &headerConfig{}
	for _, option := range options {
		option(headerConfig)
	}
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
	}
	if headerConfig.ContentDisposition != "" {
		request.ResponseContentDisposition = &headerConfig.ContentDisposition
	}

	// 构建options
	oss_options := make([]func(*oss.PresignOptions), 0)
	if headerConfig.Expires > 0 {
		oss_options = append(oss_options, oss.PresignExpires(time.Duration(headerConfig.Expires)*time.Second))
	}
	result, err := client.Presign(ctx, request, oss_options...)
	if err != nil {
		// log.Fatalf("failed to put object acl %v", err)
		return "", err
	}
	if result == nil {
		return "", fmt.Errorf("presign 设置失败, result is nil")
	}
	return result.URL, nil
}

func PushObjectCacheAdapter(ctx context.Context, accessKey, secretKey string, objectsPath []string, options ...ClientOption) (string, string, error) {
	headerConfig := &headerConfig{}
	for _, option := range options {
		option(headerConfig)
	}
	if headerConfig.PushCache {
		return PushObjectCache(ctx, accessKey, secretKey, objectsPath)
	}
	// 如果没有开启CDN缓存，则直接返回
	return "", "", nil
}

func PushObjectCache(ctx context.Context, accessKey, secretKey string, objectsPath []string, options ...PushObjectCacheOption) (string, string, error) {
	client, err := GetOSSCDNClient(ctx, accessKey, secretKey)
	if err != nil {
		return "", "", fmt.Errorf("GetOSSCDNClient failed: %v", err)
	}
	runtime := &util.RuntimeOptions{
		Autoretry:    &CdnCacheAutoRetry,
		MaxAttempts:  &CdnCacheMaxAttempts,
		KeepAlive:    &CdnCacheKeepAlive,
		MaxIdleConns: &CdnCacheMaxIdleConns,
	}
	applyPushObjectCacheOptions(options, runtime)
	urls := strings.Join(objectsPath, "\r\n")
	pushReq := &cacheClient.PushObjectCacheRequest{
		ObjectPath: tea.String(urls), // 填写需要刷新的资源路径
	}
	resp, err := client.PushObjectCacheWithOptions(pushReq, runtime)
	if err != nil {
		return "", "", err
	}
	if resp == nil {
		return "", "", fmt.Errorf("PushObjectCache failed, objectsPath: %s, result is nil", urls)
	}

	if *resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("PushObjectCache failed, objectsPath: %s, statusCode: %d", urls, *resp.StatusCode)
	}
	g.Log().Infof(ctx, "PushObjectCache success, objectsPath: %s ,PushTaskId: %s , RequestId: %s, statusCode: %d", urls, *resp.Body.PushTaskId, *resp.Body.RequestId, *resp.StatusCode)

	return *resp.Body.PushTaskId, *resp.Body.RequestId, nil
}

func applyPushObjectCacheOptions(options []PushObjectCacheOption, runtime *util.RuntimeOptions) {
	config := &pushObjectCacheConfig{}
	for _, option := range options {
		option(config)
	}
	if config.AutoRetry != nil {
		runtime.Autoretry = config.AutoRetry
	}
	if config.MaxAttempts != nil {
		runtime.MaxAttempts = config.MaxAttempts
	}
	if config.MaxIdleConns != nil {
		runtime.MaxIdleConns = config.MaxIdleConns
	}
}

func DescribeRefreshTasks(ctx context.Context, accessKey, secretKey string, taskId string) (*cacheClient.DescribeRefreshTasksResponseBody, error) {
	client, err := GetOSSCDNClient(ctx, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("GetOSSCDNClient failed: %v", err)
	}
	describeRefreshTaskByIdRequest := &cacheClient.DescribeRefreshTasksRequest{
		TaskId: tea.String(taskId), // 填写需要查询的任务ID
	}
	resp, err := client.DescribeRefreshTasks(describeRefreshTaskByIdRequest)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("DescribeRefreshTasks failed, taskId: %s, result is nil", taskId)
	}
	if *resp.StatusCode >= 400 {
		return nil, fmt.Errorf("DescribeRefreshTasks failed, taskId: %s, statusCode: %d", taskId, *resp.StatusCode)
	}
	g.Log().Infof(ctx, "DescribeRefreshTasks success, taskId: %s , RequestId: %s, statusCode: %d", taskId, *resp.Body.RequestId, *resp.StatusCode)
	return resp.Body, nil
}
