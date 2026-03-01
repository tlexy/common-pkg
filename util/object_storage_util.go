package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tlexy/common-pkg/concurrent"

	"github.com/google/uuid"
	"github.com/tlexy/common-pkg/util/acl"
	"github.com/tlexy/common-pkg/util/common"
	"github.com/tlexy/common-pkg/util/cos_util"
	obs_utils "github.com/tlexy/common-pkg/util/obs_util"
	"github.com/tlexy/common-pkg/util/oss_util"
	"github.com/tlexy/common-pkg/util/tos_util"
)

type contextKey string

const (
	contextKeyAccessKey  = contextKey("accessKey")
	contextKeySecretKey  = contextKey("secretKey")
	contextKeyEndPoint   = contextKey("endPoint")
	contextKeyBucketName = contextKey("bucketName")
	// contextKeyCdnName    = contextKey("cdnName")
	contextKeyTypeName                 = contextKey("typeName")
	contextKeyCdnName                  = contextKey("cdnName")
	contextKeyPublicEndPoint           = contextKey("publicEndPoint")
	contextKeyOssArn                   = contextKey("ossArn")
	contextKeyTosTrn                   = contextKey("tosTrn")
	ContextKeyCosUrl                   = contextKey("cosUrl")
	contextTypeObs                     = "obs"
	contextTypeCos                     = "cos"
	contextTypeOss                     = "oss"
	contextTypeTos                     = "tos"
	contextKeyPushObjectCacheAccessKey = contextKey("pushObjectCacheAccessKey")
	contextKeyPushObjectCacheSecretKey = contextKey("pushObjectCacheSecretKey")
)

var globalContext context.Context

func InitGlobalObsContext(accessKey, secretKey, endPoint, bucketName, obsCdn string) context.Context {
	ctx := BuildObsContext(accessKey, secretKey, endPoint, bucketName, obsCdn)
	globalContext = ctx
	return ctx
}

func InitGlobalOssContext(accessKey, secretKey, endPoint, bucketName, cdnUrl, arn string) context.Context {
	ctx := BuildOssContext(accessKey, secretKey, endPoint, "", bucketName, cdnUrl, arn, "", "")
	globalContext = ctx
	return ctx
}

func InitGlobalCosContext(accessKey, secretKey, cosUrl, cdnUrl string) context.Context {
	ctx := BuildCosContext(accessKey, secretKey, cosUrl, cdnUrl)
	globalContext = ctx
	return ctx
}

func BuildObsContext(accessKey, secretKey, endPoint, bucketName, obsCdn string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextKeyAccessKey, accessKey)
	ctx = context.WithValue(ctx, contextKeySecretKey, secretKey)
	ctx = context.WithValue(ctx, contextKeyEndPoint, endPoint)
	ctx = context.WithValue(ctx, contextKeyBucketName, bucketName)
	if obsCdn != "" {
		if !strings.Contains(obsCdn, "https://") {
			obsCdn = "https://" + obsCdn
		}
	}
	ctx = context.WithValue(ctx, contextKeyCdnName, obsCdn)
	ctx = context.WithValue(ctx, contextKeyTypeName, contextTypeObs)
	return ctx
}

func BuildOssContext(accessKey, secretKey, endPoint, publicEndPoint, bucketName, cdnUrl, arn, pushObjectCacheAccessKey, pushObjectCacheSecretKey string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextKeyAccessKey, accessKey)
	ctx = context.WithValue(ctx, contextKeySecretKey, secretKey)
	ctx = context.WithValue(ctx, contextKeyEndPoint, endPoint)
	ctx = context.WithValue(ctx, contextKeyBucketName, bucketName)
	ctx = context.WithValue(ctx, contextKeyOssArn, arn)
	if len(publicEndPoint) < 3 {
		publicEndPoint = endPoint
	}
	ctx = context.WithValue(ctx, contextKeyPublicEndPoint, publicEndPoint)
	if cdnUrl != "" {
		if !strings.Contains(cdnUrl, "https://") {
			cdnUrl = "https://" + cdnUrl
		}
	}
	ctx = context.WithValue(ctx, contextKeyCdnName, cdnUrl)
	ctx = context.WithValue(ctx, contextKeyTypeName, contextTypeOss)
	ctx = context.WithValue(ctx, contextKeyPushObjectCacheAccessKey, pushObjectCacheAccessKey)
	ctx = context.WithValue(ctx, contextKeyPushObjectCacheSecretKey, pushObjectCacheSecretKey)
	return ctx
}

func BuildCosContext(accessKey, secretKey, cosUrl, cdnUrl string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextKeyAccessKey, accessKey)
	ctx = context.WithValue(ctx, contextKeySecretKey, secretKey)
	ctx = context.WithValue(ctx, ContextKeyCosUrl, cosUrl)
	if cdnUrl != "" {
		if !strings.Contains(cdnUrl, "https://") {
			cdnUrl = "https://" + cdnUrl
		}
	}
	ctx = context.WithValue(ctx, contextKeyCdnName, cdnUrl)
	ctx = context.WithValue(ctx, contextKeyTypeName, contextTypeCos)

	return ctx
}

func BuildTosContext(accessKey, secretKey, endPoint, bucketName, tosCdn, tosTrn string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextKeyAccessKey, accessKey)
	ctx = context.WithValue(ctx, contextKeySecretKey, secretKey)
	ctx = context.WithValue(ctx, contextKeyEndPoint, endPoint)
	ctx = context.WithValue(ctx, contextKeyBucketName, bucketName)
	if tosCdn != "" {
		if !strings.Contains(tosCdn, "https://") {
			tosCdn = "https://" + tosCdn
		}
	}
	ctx = context.WithValue(ctx, contextKeyCdnName, tosCdn)
	ctx = context.WithValue(ctx, contextKeyTypeName, contextTypeTos)
	ctx = context.WithValue(ctx, contextKeyTosTrn, tosTrn)
	return ctx
}

func GetObsContext(ctx context.Context) (string, string, string, string, string, error) {
	accessKey, ok := ctx.Value(contextKeyAccessKey).(string)
	if !ok {
		return "", "", "", "", "", fmt.Errorf("missing access key in context")
	}
	secretKey, ok := ctx.Value(contextKeySecretKey).(string)
	if !ok {
		return "", "", "", "", "", fmt.Errorf("missing secret key in context")
	}
	endPoint, ok := ctx.Value(contextKeyEndPoint).(string)
	if !ok {
		return "", "", "", "", "", fmt.Errorf("missing end point in context")
	}
	bucketName, ok := ctx.Value(contextKeyBucketName).(string)
	if !ok {
		return "", "", "", "", "", fmt.Errorf("missing bucket name in context")
	}
	// cdn名称不一定存在
	cdnName, _ := ctx.Value(contextKeyCdnName).(string)
	return accessKey, secretKey, endPoint, bucketName, cdnName, nil
}

func GetOssContext(ctx context.Context) (string, string, string, string, string, string, error) {
	accessKey, ok := ctx.Value(contextKeyAccessKey).(string)
	if !ok {
		return "", "", "", "", "", "", fmt.Errorf("missing access key in context")
	}
	secretKey, ok := ctx.Value(contextKeySecretKey).(string)
	if !ok {
		return "", "", "", "", "", "", fmt.Errorf("missing secret key in context")
	}
	endPoint, ok := ctx.Value(contextKeyEndPoint).(string)
	if !ok {
		return "", "", "", "", "", "", fmt.Errorf("missing end point in context")
	}
	bucketName, ok := ctx.Value(contextKeyBucketName).(string)
	if !ok {
		return "", "", "", "", "", "", fmt.Errorf("missing bucket name in context")
	}
	arn, ok := ctx.Value(contextKeyOssArn).(string)
	// if!ok {
	// 	return "", "", "", "", "", "", fmt.Errorf("missing end point in context")
	// }
	// cdn名称不一定存在
	cdnName, _ := ctx.Value(contextKeyCdnName).(string)
	return accessKey, secretKey, endPoint, bucketName, cdnName, arn, nil
}

func GetOssPushObjectCacheContext(ctx context.Context) (string, string, string, error) {
	accessKey, ok := ctx.Value(contextKeyPushObjectCacheAccessKey).(string)
	if !ok {
		return "", "", "", fmt.Errorf("missing access key in context")
	}
	secretKey, ok := ctx.Value(contextKeyPushObjectCacheSecretKey).(string)
	if !ok {
		return "", "", "", fmt.Errorf("missing secret key in context")
	}

	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return "", "", "", fmt.Errorf("missing context type name")
	}

	return accessKey, secretKey, typeName, nil
}

func GetOssPublicEndPoint(ctx context.Context) (string, error) {
	publicEndPoint, ok := ctx.Value(contextKeyPublicEndPoint).(string)
	if !ok {
		return "", fmt.Errorf("missing public end point in context")
	}
	return publicEndPoint, nil
}

func GetCosContext(ctx context.Context) (string, string, string, string, error) {
	accessKey, ok := ctx.Value(contextKeyAccessKey).(string)
	if !ok {
		return "", "", "", "", fmt.Errorf("missing access key in context")
	}
	secretKey, ok := ctx.Value(contextKeySecretKey).(string)
	if !ok {
		return "", "", "", "", fmt.Errorf("missing secret key in context")
	}
	cosUrl, ok := ctx.Value(ContextKeyCosUrl).(string)
	if !ok {
		return "", "", "", "", fmt.Errorf("missing end point in context")
	}
	// cdn名称不一定存在
	cdnName, _ := ctx.Value(contextKeyCdnName).(string)
	return accessKey, secretKey, cosUrl, cdnName, nil
}

func GetTosContext(ctx context.Context) (string, string, string, string, string, string, error) {
	accessKey, ok := ctx.Value(contextKeyAccessKey).(string)
	if !ok {
		return "", "", "", "", "", "", fmt.Errorf("missing access key in context")
	}
	secretKey, ok := ctx.Value(contextKeySecretKey).(string)
	if !ok {
		return "", "", "", "", "", "", fmt.Errorf("missing secret key in context")
	}
	endPoint, ok := ctx.Value(contextKeyEndPoint).(string)
	if !ok {
		return "", "", "", "", "", "", fmt.Errorf("missing end point in context")
	}
	bucketName, ok := ctx.Value(contextKeyBucketName).(string)
	if !ok {
		return "", "", "", "", "", "", fmt.Errorf("missing bucket name in context")
	}
	tosTrn, ok := ctx.Value(contextKeyTosTrn).(string)
	if !ok {
		return "", "", "", "", "", "", fmt.Errorf("missing tos trn in context")
	}
	// cdn名称不一定存在
	cdnName, _ := ctx.Value(contextKeyCdnName).(string)
	return accessKey, secretKey, endPoint, bucketName, cdnName, tosTrn, nil
}

func extractObjectNameFromHttpUrl(remoteUrl, cdn string) (string, error) {
	pos := strings.Index(remoteUrl, cdn)
	if cdn != "" && pos > -1 {
		objectName := remoteUrl[(pos + len(cdn) + 1):]
		return eraseQuestionMark(objectName), nil
	}
	prefix := "https://"
	pos = strings.Index(remoteUrl, prefix)
	remoteUrl = remoteUrl[(pos + len(prefix) + 1):]
	pos = strings.Index(remoteUrl, "/")
	if pos > -1 {
		objectName := remoteUrl[pos+1:]
		return eraseQuestionMark(objectName), nil
	}
	return "", fmt.Errorf("missing cdn name in context")
}

func eraseQuestionMark(url string) string {
	pos := strings.Index(url, "?")
	if pos > -1 {
		return url[:pos]
	}
	return url
}

func DownloadHttpResource(ctx context.Context, remoteUrl, localFilename string) error {
	targetDir := filepath.Dir(localFilename)
	err := common.EnsureOutputDirectory(targetDir)
	if err != nil {
		return err
	}
	accessKey, secretKey, endPoint, bucketName, cdn, err := GetObsContext(ctx)
	if err != nil {
		return httpDownload(remoteUrl, localFilename)
	}
	// cdn, ok := ctx.Value(contextKeyObsCdnName).(string)
	if len(cdn) > 0 {
		objectName := ""
		pos := strings.Index(remoteUrl, cdn)
		if pos > -1 {
			objectName = remoteUrl[(pos + len(cdn) + 1):]
		}
		if objectName != "" {
			err = obs_utils.Download(ctx, accessKey, secretKey, endPoint, bucketName, objectName, localFilename)
			if err == nil {
				return nil
			}
		}
	}
	return httpDownload(remoteUrl, localFilename)
}

func GetTemporaryCredentialsWithContext(ctx context.Context, sec int64) (tmpAccessKey, tmpSecretKey, token string, err error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return "", "", "", fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		accessKey, secretKey, endPoint, _, _, err := GetObsContext(ctx)
		if err != nil {
			return "", "", "", err
		} else {
			return obs_utils.GetTemporaryCredentials(accessKey, secretKey, endPoint, int32(sec))
		}
	} else if typeName == contextTypeOss {
		accessKey, secretKey, endPoint, _, _, arn, err := GetOssContext(ctx)
		if err != nil {
			return "", "", "", err
		} else {
			return oss_util.GetTemporaryCredentials(accessKey, secretKey, endPoint, arn, int32(sec))
		}
	} else if typeName == contextTypeCos {
		accessKey, secretKey, cosUrl, _, err := GetCosContext(ctx)
		if err != nil {
			return "", "", "", err
		}
		return cos_util.GetTemporaryCredentials(ctx, accessKey, secretKey, cosUrl, sec)
	} else if typeName == contextTypeTos {
		accessKey, secretKey, endPoint, bucketName, _, trn, err := GetTosContext(ctx)
		if err != nil {
			return "", "", "", err
		}
		return tos_util.GetTemporaryCredentials(accessKey, secretKey, endPoint, bucketName, trn, sec)
	} else {
		return "", "", "", fmt.Errorf("unknown context type name")
	}
}

func GetObjectUrlWithContext(ctx context.Context, objectName string) (string, string, error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return "", "", fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		_, _, endPoint, bucketName, cdn, err := GetObsContext(ctx)
		if err != nil {
			return "", "", err
		} else {
			cdnUrl := fmt.Sprintf("%s/%s", cdn, objectName)
			return fmt.Sprintf("https://%s.%s/%s", bucketName, endPoint, objectName), cdnUrl, nil
		}
	} else if typeName == contextTypeOss {
		_, _, endPoint, bucketName, cdn, _, err := GetOssContext(ctx)
		if err != nil {
			return "", "", err
		} else {
			cdnUrl := fmt.Sprintf("%s/%s", cdn, objectName)
			return fmt.Sprintf("https://%s.%s/%s", bucketName, endPoint, objectName), cdnUrl, nil
		}
	} else if typeName == contextTypeCos {
		_, _, cosUrl, cdn, err := GetCosContext(ctx)
		if err != nil {
			return "", "", err
		} else {
			cdnUrl := fmt.Sprintf("%s/%s", cdn, objectName)
			return fmt.Sprintf("%s/%s", cosUrl, objectName), cdnUrl, nil
		}
	} else {
		return "", "", fmt.Errorf("unknown context type name")
	}
}

func DeleteObjectWithContext(ctx context.Context, objectName string) error {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		accessKey, secretKey, endPoint, bucketName, _, err := GetObsContext(ctx)
		if err != nil {
			return err
		} else {
			return obs_utils.RemoveObject(accessKey, secretKey, endPoint, bucketName, objectName)
		}
	} else if typeName == contextTypeOss {
		accessKey, secretKey, endPoint, bucketName, _, _, err := GetOssContext(ctx)
		if err != nil {
			return err
		} else {
			return oss_util.RemoveObject(accessKey, secretKey, endPoint, bucketName, objectName)
		}
	} else if typeName == contextTypeCos {
		accessKey, secretKey, cosUrl, _, err := GetCosContext(ctx)
		if err != nil {
			return err
		} else {
			return cos_util.RemoveObject(accessKey, secretKey, cosUrl, objectName)
		}
	} else {
		return fmt.Errorf("unsupported context type name")
	}
}

func GetCosTemporaryCredentialsWithContext(ctx context.Context, sec int64) (tmpAccessKey, tmpSecretKey, token string, err error) {
	accessKey, secretKey, cosUrl, _, err := GetCosContext(ctx)
	if err != nil {
		return "", "", "", err
	}
	return cos_util.GetTemporaryCredentials(ctx, accessKey, secretKey, cosUrl, sec)
}

func IsObjectExistWithContext(ctx context.Context, objectName string) (bool, error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return false, fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		accessKey, secretKey, endPoint, bucketName, _, err := GetObsContext(ctx)
		if err != nil {
			return false, err
		} else {
			return obs_utils.IsObjectExist(accessKey, secretKey, endPoint, bucketName, objectName)
		}
	} else if typeName == contextTypeOss {
		accessKey, secretKey, endPoint, bucketName, _, _, err := GetOssContext(ctx)
		if err != nil {
			return false, err
		} else {
			return oss_util.IsObjectExist(accessKey, secretKey, endPoint, bucketName, objectName)
		}
	} else if typeName == contextTypeCos {
		accessKey, secretKey, cosUrl, _, err := GetCosContext(ctx)
		if err != nil {
			return false, err
		} else {
			return cos_util.IsObjectExist(accessKey, secretKey, cosUrl, objectName)
		}
	} else {
		return false, fmt.Errorf("unsupported context type name")
	}
}

func GetObjectSizeWithContext(ctx context.Context, objectName string) (size int64, err error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return 0, fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeOss {
		accessKey, secretKey, endPoint, bucketName, _, _, err := GetOssContext(ctx)
		if err != nil {
			return -1, err
		} else {
			return oss_util.GetObjectSize(accessKey, secretKey, endPoint, bucketName, objectName)
		}
	} else {
		return -1, fmt.Errorf("unsupported context type name")
	}
}

func SetSymlinkWithContext(ctx context.Context, objectName, symlink string) (string, error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return "", fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		return "", fmt.Errorf("unsupported context type name")
	} else if typeName == contextTypeOss {
		accessKey, secretKey, endPoint, bucketName, _, _, err := GetOssContext(ctx)
		if err != nil {
			return "", err
		} else {
			return oss_util.SetSymlink(accessKey, secretKey, endPoint, bucketName, objectName, symlink)
		}
	} else if typeName == contextTypeCos {
		return "", fmt.Errorf("unsupported context type name")
	} else {
		return "", fmt.Errorf("unsupported context type name")
	}
}

// /不区分context版本
func DownloadObject(objectName, localFilename string) error {
	return DownloadObjectWithContext(globalContext, objectName, localFilename)
}

func DownloadBytes(objectName string) (io.ReadCloser, error) {
	return DownloadObjectBytesWithContext(globalContext, objectName)
}

// 上传本地文件到远端，返回远端链接以及error.
// 如果有设置了cdn地址，返回的链接是一个有效的链接，如果没有设置cdn地址，返回的链接是一个无效链接
func UploadObject(objectName, localFilename string, aclType acl.StorageAclType) (string, error) {
	return UploadObjectWithContext(globalContext, objectName, localFilename, aclType)
}

func UploadBytes(r io.ReadCloser, objectName string, aclType acl.StorageAclType) (string, string, error) {
	return UploadBytesWithContext(globalContext, r, objectName, aclType)
}

func SetPublicRead(objectName string) (string, error) {
	return SetPublicReadWithContext(globalContext, objectName)
}

func DownloadObjectWithContext(ctx context.Context, objectName, localFilename string) error {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		return downloadObsResource(ctx, objectName, localFilename)
	} else if typeName == contextTypeOss {
		return downloadOssResource(ctx, objectName, localFilename)
	} else if typeName == contextTypeCos {
		return downloadCosResource(ctx, objectName, localFilename)
	} else if typeName == contextTypeTos {
		return downloadTosResource(ctx, objectName, localFilename)
	} else {
		return fmt.Errorf("unsupported context type name")
	}
}

func DownloadObjectBytesWithContext(ctx context.Context, objectName string) (io.ReadCloser, error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return nil, fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		return downloadObsObjectBytes(ctx, objectName)
	} else if typeName == contextTypeOss {
		return downloadOssObjectBytes(ctx, objectName)
	} else if typeName == contextTypeCos {
		return downloadCosObjectBytes(ctx, objectName)
	} else if typeName == contextTypeTos {
		return downloadTosObjectBytes(ctx, objectName)
	} else {
		return nil, fmt.Errorf("unsupported context type name")
	}
}

func UploadObjectWithContext(ctx context.Context, objectName, localFilename string, aclType acl.StorageAclType, options ...oss_util.ClientOption) (string, error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return "", fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		return uploadObsResource(ctx, localFilename, objectName, aclType)
	} else if typeName == contextTypeOss {
		return uploadOssResource(ctx, localFilename, objectName, aclType, options...)
	} else if typeName == contextTypeCos {
		return uploadCosResource(ctx, localFilename, objectName, aclType)
	} else if typeName == contextTypeTos {
		return uploadTosResource(ctx, localFilename, objectName, aclType)
	} else {
		return "", fmt.Errorf("unsupported context type name")
	}
}

func PresignWithContext(ctx context.Context, objectName string, options ...oss_util.ClientOption) (string, error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return "", fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeOss {
		return presign(ctx, objectName, options...)
	} else {
		return "", fmt.Errorf("unsupported context type name")
	}
}

func UploadBytesWithContext(ctx context.Context, r io.ReadCloser, objectName string, aclType acl.StorageAclType, options ...oss_util.ClientOption) (string, string, error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return "", "", fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		return uploadObsResourceBytes(ctx, r, objectName, aclType)
	} else if typeName == contextTypeOss {
		return uploadOssResourceBytes(ctx, r, objectName, aclType, options...)
	} else if typeName == contextTypeCos {
		return uploadCosResourceBytes(ctx, r, objectName, aclType)
	} else if typeName == contextTypeTos {
		return uploadTosResourceBytes(ctx, r, objectName, aclType)
	} else {
		return "", "", fmt.Errorf("unsupported context type name")
	}
}

func SetPublicReadWithContext(ctx context.Context, objectName string) (string, error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return "", fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		return setPublicReadObsObject(ctx, objectName)
	} else if typeName == contextTypeOss {
		return setPublicReadOssObject(ctx, objectName)
	} else if typeName == contextTypeCos {
		return setPublicReadCosObject(ctx, objectName)
	} else {
		return "", fmt.Errorf("unsupported context type name")
	}
}

func SetPrivateWithContext(ctx context.Context, objectName string) (string, error) {
	typeName, ok := ctx.Value(contextKeyTypeName).(string)
	if !ok {
		return "", fmt.Errorf("missing context type name")
	}
	if typeName == contextTypeObs {
		return "", fmt.Errorf("obs not support")
	} else if typeName == contextTypeOss {
		return setPrivateOssObject(ctx, objectName)
	} else if typeName == contextTypeCos {
		return "", fmt.Errorf("cos not support")
	} else {
		return "", fmt.Errorf("unsupported context type name")
	}
}

func downloadObsResource(ctx context.Context, objectName, localFilename string) error {
	targetDir := filepath.Dir(localFilename)
	err := common.EnsureOutputDirectory(targetDir)
	if err != nil {
		return err
	}
	accessKey, secretKey, endPoint, bucketName, cdn, err := GetObsContext(ctx)
	if err != nil {
		return fmt.Errorf("missing obs context")
	}
	if strings.Contains(objectName, "https://") {
		objectName, err = extractObjectNameFromHttpUrl(objectName, cdn)
		if err != nil {
			return err
		}
	}
	return obs_utils.Download(ctx, accessKey, secretKey, endPoint, bucketName, objectName, localFilename)
}

func downloadOssResource(ctx context.Context, objectName, localFilename string) error {
	targetDir := filepath.Dir(localFilename)
	err := common.EnsureOutputDirectory(targetDir)
	if err != nil {
		return err
	}
	accessKey, secretKey, endPoint, bucketName, cdn, _, err := GetOssContext(ctx)
	if err != nil {
		return fmt.Errorf("missing oss context")
	}
	if strings.Contains(objectName, "https://") {
		objectName, err = extractObjectNameFromHttpUrl(objectName, cdn)
		if err != nil {
			return err
		}
	}
	return oss_util.Download(ctx, accessKey, secretKey, endPoint, bucketName, objectName, localFilename)
}

func downloadCosResource(ctx context.Context, objectName, localFilename string) error {
	targetDir := filepath.Dir(localFilename)
	err := common.EnsureOutputDirectory(targetDir)
	if err != nil {
		return err
	}
	accessKey, secretKey, cosUrl, cdn, err := GetCosContext(ctx)
	if err != nil {
		return fmt.Errorf("missing oss context")
	}
	if strings.Contains(objectName, "https://") {
		objectName, err = extractObjectNameFromHttpUrl(objectName, cdn)
		if err != nil {
			return err
		}
	}
	return cos_util.Download(ctx, accessKey, secretKey, cosUrl, objectName, localFilename)
}

func downloadTosResource(ctx context.Context, objectName, localFilename string) error {
	targetDir := filepath.Dir(localFilename)
	err := common.EnsureOutputDirectory(targetDir)
	if err != nil {
		return err
	}
	accessKey, secretKey, endPoint, bucketName, cdn, _, err := GetTosContext(ctx)
	if err != nil {
		return fmt.Errorf("missing tos context")
	}
	if strings.Contains(objectName, "https://") {
		objectName, err = extractObjectNameFromHttpUrl(objectName, cdn)
		if err != nil {
			return err
		}
	}
	return tos_util.Download(ctx, accessKey, secretKey, endPoint, bucketName, objectName, localFilename)
}

func downloadObsObjectBytes(ctx context.Context, objectName string) (io.ReadCloser, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, err := GetObsContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("missing obs context")
	}
	if strings.Contains(objectName, "https://") {
		objectName, err = extractObjectNameFromHttpUrl(objectName, cdn)
		if err != nil {
			return nil, err
		}
	}
	return obs_utils.DownloadBytes(ctx, accessKey, secretKey, endPoint, bucketName, objectName)
}

func downloadOssObjectBytes(ctx context.Context, objectName string) (io.ReadCloser, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, _, err := GetOssContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("missing obs context")
	}
	if strings.Contains(objectName, "https://") {
		objectName, err = extractObjectNameFromHttpUrl(objectName, cdn)
		if err != nil {
			return nil, err
		}
	}
	return oss_util.DownloadBytes(ctx, accessKey, secretKey, endPoint, bucketName, objectName)
}

func downloadCosObjectBytes(ctx context.Context, objectName string) (io.ReadCloser, error) {
	accessKey, secretKey, cosUrl, cdn, err := GetCosContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("missing cos context")
	}
	if strings.Contains(objectName, "https://") {
		objectName, err = extractObjectNameFromHttpUrl(objectName, cdn)
		if err != nil {
			return nil, err
		}
	}
	return cos_util.DownloadBytes(ctx, accessKey, secretKey, cosUrl, objectName)
}

func downloadTosObjectBytes(ctx context.Context, objectName string) (io.ReadCloser, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, _, err := GetTosContext(ctx)
	if err != nil {
		return nil, err
	}
	if strings.Contains(objectName, "https://") {
		objectName, err = extractObjectNameFromHttpUrl(objectName, cdn)
		if err != nil {
			return nil, err
		}
	}
	return tos_util.DownloadBytes(ctx, accessKey, secretKey, endPoint, bucketName, objectName)
}

func uploadObsResource(ctx context.Context, localFilename, objectName string, aclType acl.StorageAclType) (string, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, err := GetObsContext(ctx)
	if err != nil {
		return "", fmt.Errorf("missing obs context")
	}
	// 获取aux目录
	auxDir := filepath.Dir(localFilename)
	auxDir = auxDir + "/upload_big_aux_" + uuid.New().String()
	err = obs_utils.ObsUpload(accessKey, secretKey, endPoint, bucketName, objectName, localFilename, aclType, auxDir)
	if err != nil {
		return "", err
	}
	if cdn == "" {
		cdn = fmt.Sprintf("https://%s.%s", bucketName, endPoint)
	}
	return fmt.Sprintf("%s/%s", cdn, objectName), nil
	// prefix := fmt.Sprintf("https://%s.%s", bucketName, endPoint)
	// directUrl := fmt.Sprintf("%s/%s", prefix, objectName)
	// if cdn == "" {
	// 	return "", directUrl, nil
	// } else {
	// 	return fmt.Sprintf("%s/%s", cdn, objectName), directUrl, nil
	// }
}

func presign(ctx context.Context, objectName string, options ...oss_util.ClientOption) (string, error) {
	accessKey, secretKey, _, bucketName, _, _, err := GetOssContext(ctx)
	if err != nil {
		return "", err
	}
	publcEndPoint, err := GetOssPublicEndPoint(ctx)
	if err != nil || publcEndPoint == "" {
		return "", fmt.Errorf("missing public end point")
	}
	return oss_util.Presign(ctx, accessKey, secretKey, publcEndPoint, bucketName, objectName, options...)
}

func uploadOssResource(ctx context.Context, localFilename, objectName string, aclType acl.StorageAclType, options ...oss_util.ClientOption) (string, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, _, err := GetOssContext(ctx)
	if err != nil {
		return "", fmt.Errorf("missing oss context")
	}
	// 获取aux目录
	// auxDir := filepath.Dir(localFilename)
	// auxDir = auxDir + "/upload_big_aux_" + uuid.New().String()
	err = oss_util.Upload(ctx, accessKey, secretKey, endPoint, bucketName, localFilename, objectName, aclType, options...)
	if err != nil {
		return "", err
	}

	// prefix := fmt.Sprintf("https://%s.%s", bucketName, endPoint)
	// directUrl := fmt.Sprintf("%s/%s", prefix, objectName)
	// if cdn == "" {
	// 	return "", directUrl, nil
	// } else {
	// 	return fmt.Sprintf("%s/%s", cdn, objectName), directUrl, nil
	// }
	needPushCache := cdn != ""
	if cdn == "" {
		cdn = fmt.Sprintf("https://%s.%s", bucketName, endPoint)
	}
	url := fmt.Sprintf("%s/%s", cdn, objectName)
	if needPushCache {
		doPushObjectCache(ctx, url, options...)
	}
	return url, nil
}

func doPushObjectCache(ctx context.Context, url string, options ...oss_util.ClientOption) {
	concurrent.GoSafe(func() {
		accessKey, secretKey, _, err := GetOssPushObjectCacheContext(ctx)
		if err != nil {
			return
		}
		if accessKey == "" || secretKey == "" {
			return
		}
		oss_util.PushObjectCacheAdapter(ctx, accessKey, secretKey, []string{url}, options...)
	})
}

func uploadCosResource(ctx context.Context, localFilename, objectName string, aclType acl.StorageAclType) (string, error) {
	accessKey, secretKey, cosUrl, cdn, err := GetCosContext(ctx)
	if err != nil {
		return "", fmt.Errorf("missing cos context")
	}
	err = cos_util.Upload(ctx, accessKey, secretKey, cosUrl, objectName, localFilename, aclType)
	if err != nil {
		return "", err
	}
	if cdn == "" {
		cdn = cosUrl
	}
	return fmt.Sprintf("%s/%s", cdn, objectName), nil
}

func uploadTosResource(ctx context.Context, localFilename, objectName string, aclType acl.StorageAclType) (string, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, _, err := GetTosContext(ctx)
	if err != nil {
		return "", fmt.Errorf("missing tos context")
	}
	err = tos_util.Upload(ctx, accessKey, secretKey, endPoint, bucketName, localFilename, objectName, aclType)
	if err != nil {
		return "", err
	}
	if cdn == "" {
		cdn = fmt.Sprintf("https://%s.%s", bucketName, strings.Replace(endPoint, "https://", "", -1))
	}
	return fmt.Sprintf("%s/%s", cdn, objectName), nil
}

func uploadObsResourceBytes(ctx context.Context, r io.ReadCloser, objectName string, aclType acl.StorageAclType) (string, string, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, err := GetObsContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("missing obs context")
	}
	// 获取aux目录
	// auxDir := "/upload_big_aux_" + uuid.New().String()
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return "", "", err
	}
	err = obs_utils.UploadBytes(accessKey, secretKey, endPoint, bucketName, objectName, data, aclType)
	if err != nil {
		return "", "", err
	}
	prefix := fmt.Sprintf("https://%s.%s", bucketName, endPoint)
	directUrl := fmt.Sprintf("%s/%s", prefix, objectName)
	if cdn == "" {
		return "", directUrl, nil
	} else {
		return fmt.Sprintf("%s/%s", cdn, objectName), directUrl, nil
	}
}

func uploadOssResourceBytes(ctx context.Context, r io.ReadCloser, objectName string, aclType acl.StorageAclType, options ...oss_util.ClientOption) (string, string, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, err := GetObsContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("missing obs context")
	}
	// 获取aux目录
	// auxDir := "/upload_big_aux_" + uuid.New().String()
	defer r.Close()
	// data, err := io.ReadAll(r)
	// if err != nil {
	// 	return "", err
	// }
	err = oss_util.UploadBytesByReader(ctx, accessKey, secretKey, endPoint, bucketName, r, objectName, aclType, options...)
	if err != nil {
		return "", "", err
	}
	prefix := fmt.Sprintf("https://%s.%s", bucketName, endPoint)
	directUrl := fmt.Sprintf("%s/%s", prefix, objectName)
	if cdn == "" {
		return "", directUrl, nil
	} else {
		// 预热
		url := fmt.Sprintf("%s/%s", cdn, objectName)
		doPushObjectCache(ctx, url, options...)
		return url, directUrl, nil
	}
}

func uploadCosResourceBytes(ctx context.Context, r io.ReadCloser, objectName string, aclType acl.StorageAclType) (string, string, error) {
	accessKey, secretKey, cosUrl, cdn, err := GetCosContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("missing cos context")
	}
	defer r.Close()
	err = cos_util.UploadBytesByReader(ctx, accessKey, secretKey, cosUrl, objectName, r, aclType)
	if err != nil {
		return "", "", err
	}
	directUrl := fmt.Sprintf("%s/%s", cosUrl, objectName)
	if cdn == "" {
		return "", directUrl, nil
	} else {
		return fmt.Sprintf("%s/%s", cdn, objectName), directUrl, nil
	}
}

func uploadTosResourceBytes(ctx context.Context, r io.ReadCloser, objectName string, aclType acl.StorageAclType) (string, string, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, _, err := GetTosContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("missing tos context")
	}
	defer r.Close()
	err = tos_util.UploadBytesByReader(ctx, accessKey, secretKey, endPoint, bucketName, r, objectName, aclType)
	if err != nil {
		return "", "", err
	}
	directUrl := fmt.Sprintf("https://%s.%s/%s", bucketName, strings.Replace(endPoint, "https://", "", -1), objectName)
	if cdn == "" {
		return "", directUrl, nil
	} else {
		return fmt.Sprintf("%s/%s", cdn, objectName), directUrl, nil
	}
}

func httpDownload(remoteUrl, localFilename string) error {
	// 使用http下载文件
	resp, err := http.Get(remoteUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = os.WriteFile(localFilename, body, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func setPublicReadObsObject(ctx context.Context, objectName string) (string, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, err := GetObsContext(ctx)
	if err != nil {
		return "", fmt.Errorf("missing obs context")
	}
	err = obs_utils.SetPublic(accessKey, secretKey, endPoint, bucketName, objectName)
	if err != nil {
		return "", err
	}
	// 返回public url
	if cdn == "" {
		cdn = fmt.Sprintf("https://%s.%s", bucketName, endPoint)
	}
	return fmt.Sprintf("%s/%s", cdn, objectName), nil
}

func setPublicReadCosObject(ctx context.Context, objectName string) (string, error) {
	accessKey, secretKey, cosUrl, cdn, err := GetCosContext(ctx)
	if err != nil {
		return "", fmt.Errorf("missing obs context")
	}
	err = cos_util.SetPublic(ctx, accessKey, secretKey, cosUrl, objectName)
	if err != nil {
		return "", err
	}
	// 返回public url
	if cdn == "" {
		cdn = cosUrl
	}
	return fmt.Sprintf("%s/%s", cdn, objectName), nil
}

func setPublicReadOssObject(ctx context.Context, objectName string) (string, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, _, err := GetOssContext(ctx)
	if err != nil {
		return "", fmt.Errorf("missing obs context")
	}
	err = oss_util.SetPublic(ctx, accessKey, secretKey, endPoint, bucketName, objectName)
	if err != nil {
		return "", err
	}
	// 返回public url
	if cdn == "" {
		cdn = fmt.Sprintf("https://%s.%s", bucketName, endPoint)
	}
	return fmt.Sprintf("%s/%s", cdn, objectName), nil
}

func setPrivateOssObject(ctx context.Context, objectName string) (string, error) {
	accessKey, secretKey, endPoint, bucketName, cdn, _, err := GetOssContext(ctx)
	if err != nil {
		return "", fmt.Errorf("missing obs context")
	}
	err = oss_util.SetPrivate(ctx, accessKey, secretKey, endPoint, bucketName, objectName)
	if err != nil {
		return "", err
	}
	//返回public url
	if cdn == "" {
		cdn = fmt.Sprintf("https://%s.%s", bucketName, endPoint)
	}
	return fmt.Sprintf("%s/%s", cdn, objectName), nil
}

func PushObjectCacheWithContext(ctx context.Context, urls []string, options ...oss_util.PushObjectCacheOption) (string, string, error) {
	accessKey, secretKey, typeName, err := GetOssPushObjectCacheContext(ctx)
	if err != nil {
		return "", "", fmt.Errorf("missing push object cache context")
	}

	if typeName != contextTypeOss {
		return "", "", fmt.Errorf("unsupported context type name for push object cache: %s", typeName)
	}

	return oss_util.PushObjectCache(ctx, accessKey, secretKey, urls, options...)
}

func GetObjectNameByHttpsUrl(httpsUrl string) (string, error) {

	//兼容http协议以及带签名的链接
	if strings.HasPrefix(httpsUrl, "http://") {
		httpsUrl = strings.Replace(httpsUrl, "http://", "https://", 1)
	}
	prefix := "https://"
	pos := strings.Index(httpsUrl, prefix)
	if pos == -1 {
		return "", fmt.Errorf("url is not https url")
	}
	str := httpsUrl[pos+len(prefix)+1:]
	pos = strings.Index(str, "/")
	if pos == -1 {
		return "", fmt.Errorf("object name is not in url")
	}
	objectName := str[pos+1:]
	pos3 := strings.Index(objectName, "?")
	if pos3 == -1 {
		return objectName, nil
	}
	//去掉？号后面的内容
	return objectName[:pos3], nil
}

func ReplaceUrlToCdn(remoteUrl string, cdn string) string {
	// 替换为cdn地址
	pos := strings.Index(remoteUrl, cdn)
	if pos > -1 {
		return remoteUrl
	}
	key1 := "://"
	pos = strings.Index(remoteUrl, key1)
	if pos == -1 {
		return remoteUrl
	}
	if len(remoteUrl) < pos+len(key1)+2 {
		return remoteUrl
	}
	substr := remoteUrl[pos+len(key1)+2:]
	pos = strings.Index(substr, "/")
	if pos == -1 {
		return remoteUrl
	}
	substr = substr[pos:]
	return fmt.Sprintf("https://%s%s", cdn, substr)
}
