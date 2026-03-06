package utils

import (
	"testing"

	"github.com/tlexy/common-pkg/util/acl"
)

func TestTos_GetTemporaryCredentials2(t *testing.T) {
	storage, err := NewObjectStorage(&StorageParameter{
		AccessKey:  "",
		SecretKey:  "==",
		EndPoint:   "",
		BucketName: "",
		TosTrn:     "",
		CdnName:    "",
	}, StorageTypeTos)
	if err != nil {
		t.Fatal(err)
	}

	//73880754 2120526407

	accessKey, secretKey, sessionToken, err := storage.GetTemporaryCredentials(3600)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("accessKey: %s, secretKey: %s, sessionToken: %s", accessKey, secretKey, sessionToken)
}

func TestTos_UploadFile2(t *testing.T) {
	storage, err := NewObjectStorage(&StorageParameter{
		AccessKey:  "",
		SecretKey:  "==",
		EndPoint:   "",
		BucketName: "",
		TosTrn:     "",
		CdnName:    "",
	}, StorageTypeTos)
	if err != nil {
		t.Fatal(err)
	}

	objectName := "translate-saas/public/test.txt"
	remoteUrl, err := storage.UploadFile("object_storage_util.go", objectName, acl.AclTypePublicRead)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(remoteUrl)
	}
}

func TestTos_DownloadFile2(t *testing.T) {
	storage, err := NewObjectStorage(&StorageParameter{
		AccessKey:  "",
		SecretKey:  "==",
		EndPoint:   "",
		BucketName: "",
		TosTrn:     "",
		CdnName:    "",
	}, StorageTypeTos)
	if err != nil {
		t.Fatal(err)
	}

	objectName := "translate-saas/public/test.txt"
	localPath := "test.txt"
	err = storage.DownloadFile(objectName, localPath)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("File downloaded successfully: %s", localPath)
	}
}
