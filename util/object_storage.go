package utils

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/tlexy/common-pkg/util/oss_util"

	"github.com/tlexy/common-pkg/util/acl"
)

type StorageParameter struct {
	AccessKey                string
	SecretKey                string
	EndPoint                 string
	PublicEndPoint           string
	BucketName               string
	CosUrl                   string
	CdnName                  string
	OssArn                   string
	TosTrn                  string // 字节火山云trn
	PushObjectCacheAccessKey string // cdn预热key
	PushObjectCacheSecretKey string // cdn预热secret
}

type ObjectStorage struct {
	param *StorageParameter
	ctx   context.Context
}

type ObjecetStorageType int

const (
	StorageTypeObs ObjecetStorageType = iota // 华为云
	StorageTypeOss                           // 阿里云
	StorageTypeCos                           // 腾讯云
	StorageTypeTos                           // 字节火山云
)

func NewObjectStorage(arg *StorageParameter, ctype ObjecetStorageType) (*ObjectStorage, error) {
	ooc := &ObjectStorage{
		param: arg,
	}
	if ctype == StorageTypeOss {
		if len(arg.OssArn) < 3 {
			return nil, fmt.Errorf("oss arn is empty, if don't need to call GetTemporaryCredentials, set arn to arbitary value(more then 3 char)")
		}
		ooc.ctx = BuildOssContext(ooc.param.AccessKey, ooc.param.SecretKey, ooc.param.EndPoint, ooc.param.PublicEndPoint,
			ooc.param.BucketName, ooc.param.CdnName, arg.OssArn, ooc.param.PushObjectCacheAccessKey, ooc.param.PushObjectCacheSecretKey)
		return ooc, nil
	} else if ctype == StorageTypeObs {
		ooc.ctx = BuildObsContext(ooc.param.AccessKey, ooc.param.SecretKey, ooc.param.EndPoint, ooc.param.BucketName, ooc.param.CdnName)
		return ooc, nil
	} else if ctype == StorageTypeCos {
		if arg.CosUrl == "" {
			return nil, fmt.Errorf("cos url is empty")
		}
		ooc.ctx = BuildCosContext(ooc.param.AccessKey, ooc.param.SecretKey, ooc.param.CosUrl, ooc.param.CdnName)
		return ooc, nil
	} else if ctype == StorageTypeTos {
		ooc.ctx = BuildTosContext(ooc.param.AccessKey, ooc.param.SecretKey, ooc.param.EndPoint, ooc.param.BucketName, ooc.param.CdnName, ooc.param.TosTrn)
		return ooc, nil
	} else {
		return nil, fmt.Errorf("unsupport storage type")
	}
}

func (obs *ObjectStorage) GetCosTemporaryCredentials(sec int64) (accessKey, secretKey, token string, err error) {
	return GetCosTemporaryCredentialsWithContext(obs.ctx, sec)
}

func (obs *ObjectStorage) GetTemporaryCredentials(sec int64) (accessKey, secretKey, token string, err error) {
	return GetTemporaryCredentialsWithContext(obs.ctx, sec)
}

func (obs *ObjectStorage) IsObjectExist(objectName string) (bool, error) {
	return IsObjectExistWithContext(obs.ctx, objectName)
}

func (obs *ObjectStorage) SetSymlink(objectName, symlink string) (string, error) {
	return SetSymlinkWithContext(obs.ctx, objectName, symlink)
}

func (obs *ObjectStorage) DownloadFile(objectName, localPath string) error {
	return DownloadObjectWithContext(obs.ctx, objectName, localPath)
}

func (obs *ObjectStorage) DownloadBytes(objectName string) (io.ReadCloser, error) {
	return DownloadObjectBytesWithContext(obs.ctx, objectName)
}

func (obs *ObjectStorage) GetObjectSize(objectName string) (int64, error) {
	return GetObjectSizeWithContext(obs.ctx, objectName)
}

// 返回https链接及错误，如果对象是私有的，那么返回的链接只有设置对象为公开后才可以访问
func (obs *ObjectStorage) UploadFile(localPath, objectName string, aclType acl.StorageAclType, options ...oss_util.ClientOption) (string, error) {
	return UploadObjectWithContext(obs.ctx, objectName, localPath, aclType, options...)
}

func (obs *ObjectStorage) UploadBytes(r io.ReadCloser, objectName string, aclType acl.StorageAclType, options ...oss_util.ClientOption) (string, string, error) {
	return UploadBytesWithContext(obs.ctx, r, objectName, aclType, options...)
}

// 返回https链接及错误
func (obs *ObjectStorage) SetPublicRead(objectName string) (string, error) {
	return SetPublicReadWithContext(obs.ctx, objectName)
}

func (obs *ObjectStorage) SetPrivate(objectName string) (string, error) {
	return SetPrivateWithContext(obs.ctx, objectName)
}

func (obs *ObjectStorage) GetObjectUrl(objectName string) (string, string, error) {
	return GetObjectUrlWithContext(obs.ctx, objectName)
}

func (obs *ObjectStorage) DeleteObject(objectName string) error {
	return DeleteObjectWithContext(obs.ctx, objectName)
}

func (obs *ObjectStorage) Presign(objectName string, options ...oss_util.ClientOption) (string, error) {
	return PresignWithContext(obs.ctx, objectName, options...)
}

func (obs *ObjectStorage) PushObjectCache(cndUrls []string, options ...oss_util.PushObjectCacheOption) (string, string, error) {
	return PushObjectCacheWithContext(obs.ctx, cndUrls, options...)
}

// 通过对象名获取内网链接
func (obs *ObjectStorage) GetInternalUrl(objectName string) (string, error) {
	internalUrl, _, err := obs.GetObjectUrl(objectName)
	if err != nil {
		return "", fmt.Errorf("获取内部URL失败: %w", err)
	}
	return internalUrl, nil
}

// 通过https链接获取内网链接
func (obs *ObjectStorage) GetInternalUrlByHttps(link string) (string, error) {
	objectName, err := GetObjectNameByHttpsUrl(link)
	if err != nil {
		return "", fmt.Errorf("获取对象名失败: %w", err)
	}
	return obs.GetInternalUrl(objectName)
}

// 通过对象名获取内网签名链接
func (obs *ObjectStorage) GetInternalPresign(objectName string, options ...oss_util.ClientOption) (string, error) {
	presignUrl, err := obs.Presign(objectName, options...)
	if err != nil {
		return "", err
	}
	if len(obs.param.EndPoint) > 3 && len(obs.param.PublicEndPoint) > 3 {
		presignUrl = strings.Replace(presignUrl, obs.param.PublicEndPoint, obs.param.EndPoint, 1)
	}
	return presignUrl, nil
}

// 通过https链接获取内网签名链接
func (obs *ObjectStorage) GetInternalPresignByHttps(link string) (string, error) {
	objectName, err := GetObjectNameByHttpsUrl(link)
	if err != nil {
		return "", fmt.Errorf("获取对象名失败: %w", err)
	}
	return obs.GetInternalPresign(objectName)
}

// 返回cdn的签名链接
func (obs *ObjectStorage) PresignUsingCdn(objectName string, options ...oss_util.ClientOption) (string, error) {
	presignUrl, err := obs.Presign(objectName, options...)
	if err != nil {
		return "", err
	}
	return ReplaceUrlToCdn(presignUrl, obs.param.CdnName), nil
}

func (obs *ObjectStorage) PresignUsingCdnByHttps(link string, options ...oss_util.ClientOption) (string, error) {
	objectName, err := GetObjectNameByHttpsUrl(link)
	if err != nil {
		return "", fmt.Errorf("获取对象名失败: %w", err)
	}
	return obs.PresignUsingCdn(objectName, options...)
}
