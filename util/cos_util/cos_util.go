package cos_util

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"github.com/tlexy/common-pkg/util/acl"
)

const (
	CosChunkSize int64 = 1024 * 3
)

// 下载cos对象到本地
func Download(ctx context.Context, accessKey, secretKey, cosUrl string, objectName string, localPath string) error {
	// 将 examplebucket-1250000000 和 COS_REGION 修改为用户真实的信息
	// 存储桶名称，由 bucketname-appid 组成，appid 必须填入，可以在 COS 控制台查看存储桶名称。https://console.cloud.tencent.com/cos5/bucket
	// COS_REGION 可以在控制台查看，https://console.cloud.tencent.com/cos5/bucket, 关于地域的详情见 https://cloud.tencent.com/document/product/436/6224
	u, err := url.Parse(cosUrl)
	if err != nil {
		return err
	}

	b := &cos.BaseURL{BucketURL: u}
	// 1.永久密钥
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessKey,
			SecretKey: secretKey,
		},
	})
	// .获取对象到本地文件
	_, err = client.Object.GetToFile(context.Background(), objectName, localPath, nil)
	if err != nil {
		return err
	}

	return nil
}

func getRegionByCosUrl(cosUrl string) string {
	key := ".cos."
	pos := strings.Index(cosUrl, key)
	if pos == -1 {
		return ""
	}
	str := cosUrl[pos+len(key):]
	pos2 := strings.Index(str, ".")
	if pos2 == -1 {
		return ""
	}
	return str[:pos2]
}

func GetTemporaryCredentials(ctx context.Context, accessKey, secretKey, cosUrl string, sec int64) (tmpAccessKey, tmpSecretKey, token string, err error) {
	c := sts.NewClient(
		accessKey, // 用户的 SecretId，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参考https://cloud.tencent.com/document/product/598/37140
		secretKey, // 用户的 SecretKey，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参考https://cloud.tencent.com/document/product/598/37140
		nil,
		// sts.Host("sts.internal.tencentcloudapi.com"), // 设置域名, 默认域名sts.tencentcloudapi.com
		// sts.Scheme("http"),      // 设置协议, 默认为https，公有云sts获取临时密钥不允许走http，特殊场景才需要设置http
	)
	region := getRegionByCosUrl(cosUrl)
	// 策略概述 https://cloud.tencent.com/document/product/436/18023
	opt := &sts.CredentialOptions{
		DurationSeconds: sec,
		Region:          region,
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{
				{
					// 密钥的权限列表。简单上传和分片需要以下的权限，其他权限列表请看 https://cloud.tencent.com/document/product/436/31923
					Action: []string{
						// 简单上传
						"name/cos:PostObject",
						"name/cos:PutObject",
						// 分片上传
						"name/cos:InitiateMultipartUpload",
						"name/cos:ListMultipartUploads",
						"name/cos:ListParts",
						"name/cos:UploadPart",
						"name/cos:CompleteMultipartUpload",
					},
					Effect: "allow",
					Resource: []string{
						// 这里改成允许的路径前缀，可以根据自己网站的用户登录态判断允许上传的具体路径，例子： a.jpg 或者 a/* 或者 * (使用通配符*存在重大安全风险, 请谨慎评估使用)
						// 存储桶的命名格式为 BucketName-APPID，此处填写的 bucket 必须为此格式
						"*",
					},
					// 开始构建生效条件 condition
					// 关于 condition 的详细设置规则和COS支持的condition类型可以参考https://cloud.tencent.com/document/product/436/71306
					/*Condition: map[string]map[string]interface{}{
						"ip_equal": map[string]interface{}{
							"qcs:ip": []string{
								"",
							},
						},
					},*/
				},
			},
		},
	}

	// case 1 请求临时密钥
	res, err := c.GetCredential(opt)
	if err != nil {
		// panic(err)
		return "", "", "", err
	}
	if res.Credentials == nil {
		return "", "", "", fmt.Errorf("res.Credentials is nil")
	}
	// 临时密钥
	return res.Credentials.TmpSecretID, res.Credentials.TmpSecretKey, res.Credentials.SessionToken, nil
	// fmt.Printf("%+v\n", res)
	// fmt.Printf("%+v\n", res.Credentials)
}

func IsObjectExist(accessKey, secretKey, cosUrl string, objectName string) (bool, error) {
	u, err := url.Parse(cosUrl)
	if err != nil {
		return false, err
	}
	b := &cos.BaseURL{BucketURL: u}
	// 1.永久密钥
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessKey,
			SecretKey: secretKey,
		},
	})

	flag, err := client.Object.IsExist(context.Background(), objectName)
	if err != nil {
		return false, nil
	}
	return flag, nil
}

func RemoveObject(accessKey, secretKey, cosUrl string, objectName string) error {
	u, err := url.Parse(cosUrl)
	if err != nil {
		return err
	}
	b := &cos.BaseURL{BucketURL: u}
	// 1.永久密钥
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessKey,
			SecretKey: secretKey,
		},
	})
	_, err = client.Object.Delete(context.Background(), objectName)
	if err != nil {
		return err
	}
	return nil
}

func DownloadBytes(ctx context.Context, accessKey, secretKey, cosUrl string, objectName string) (io.ReadCloser, error) {
	u, err := url.Parse(cosUrl)
	if err != nil {
		return nil, err
	}

	b := &cos.BaseURL{BucketURL: u}
	// 1.永久密钥
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessKey,
			SecretKey: secretKey,
		},
	})
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	resp, err := client.Object.Get(context.Background(), objectName, nil)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("resp is nil")
	}
	if resp.Body == nil {
		return nil, fmt.Errorf("resp.Body is nil")
	}
	return resp.Body, nil
}

func Upload(ctx context.Context, accessKey, secretKey, cosUrl string, objectName string, localPath string, aclType acl.StorageAclType) error {
	fi, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("上传时找不到文件， file: %s", localPath)
	}

	if fi.Size() < 1024*1024*CosChunkSize {
		// 小于1M，直接上传
		return uploadSmallFile(ctx, accessKey, secretKey, cosUrl, objectName, localPath, aclType)
	} else {
		// 大于1M，分块上传
		return fmt.Errorf("文件过大，暂不支持")
	}
}

func uploadSmallFile(ctx context.Context, accessKey, secretKey, cosUrl string, objectName string, localPath string, aclType acl.StorageAclType) error {
	u, err := url.Parse(cosUrl)
	if err != nil {
		return err
	}

	b := &cos.BaseURL{BucketURL: u}
	// 1.永久密钥
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessKey,
			SecretKey: secretKey,
		},
	})

	aclTypeStr := cos.ACL.Default
	if aclType == acl.AclTypePublicRead {
		aclTypeStr = cos.ACL.PublicRead
	}
	opt := &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			// 如果不是必要操作，建议上传文件时不要给单个文件设置权限，避免达到限制。若不设置默认继承桶的权限。
			XCosACL: aclTypeStr,
		},
	}

	_, err = client.Object.PutFromFile(context.Background(), objectName, localPath, opt)
	if err != nil {
		return nil
	}
	return nil
}

func UploadBytesByReader(ctx context.Context, accessKey, secretKey, cosUrl string, objectName string, reader io.Reader, aclType acl.StorageAclType) error {
	u, err := url.Parse(cosUrl)
	if err != nil {
		return err
	}

	b := &cos.BaseURL{BucketURL: u}
	// 1.永久密钥
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessKey,
			SecretKey: secretKey,
		},
	})

	aclTypeStr := cos.ACL.Default
	if aclType == acl.AclTypePublicRead {
		aclTypeStr = cos.ACL.PublicRead
	}
	opt := &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			// 如果不是必要操作，建议上传文件时不要给单个文件设置权限，避免达到限制。若不设置默认继承桶的权限。
			XCosACL: aclTypeStr,
		},
	}

	_, err = client.Object.Put(context.Background(), objectName, reader, opt)
	return err
}

func SetPublic(ctx context.Context, accessKey, secretKey, cosUrl string, objectName string) error {
	u, err := url.Parse(cosUrl)
	if err != nil {
		return err
	}

	b := &cos.BaseURL{BucketURL: u}
	// 1.永久密钥
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessKey,
			SecretKey: secretKey,
		},
	})

	// 1.通过请求头设置
	opt := &cos.ObjectPutACLOptions{
		Header: &cos.ACLHeaderOptions{
			XCosACL: "public-read",
		},
	}
	_, err = client.Object.PutACL(context.Background(), objectName, opt)
	if err != nil {
		return err
	}
	return nil
}
