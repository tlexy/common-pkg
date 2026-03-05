package volce_base

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type VodV2Session struct {
	accessKey string
	secretKey string
	addr      string
	action    string
	version   string
	method    string
	path      string
	service   string
	region    string
}

func NewVodV2Session(accessKey, secretKey string) *VodV2Session {
	return &VodV2Session{
		accessKey: accessKey,
		secretKey: secretKey,
		addr:      "https://vod.volcengineapi.com",
		action:    "StartExecution",
		version:   "2025-01-01",
		method:    http.MethodPost,
		path:      "/",
		service:   "vod",
		region:    "cn-north-1",
	}
}

func hashSHA256(data []byte) []byte {
	hash := sha256.New()
	if _, err := hash.Write(data); err != nil {
		log.Printf("input hash err:%s", err.Error())
	}

	return hash.Sum(nil)
}

func hmacSHA256(key []byte, content string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

func getSignedKey(secretKey, date, region, service string) []byte {
	kDate := hmacSHA256([]byte(secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "request")

	return kSigning
}

func (v *VodV2Session) ConstructHttpRequest(body []byte) (*http.Request, error) {
	// https://www.volcengine.com/docs/4/1923688?lang=zh
	// https://github.com/volcengine/volc-openapi-demos/blob/main/signature/golang/sign.go
	queries := make(url.Values)
	queries.Set("Action", v.action)
	queries.Set("Version", v.version)

	requestAddr := fmt.Sprintf("%s%s?%s", v.addr, v.path, queries.Encode())
	log.Printf("request addr: %s\n", requestAddr)

	request, err := http.NewRequest(v.method, requestAddr, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("bad request: %w", err)
	}

	// 2. 构建签名材料
	now := time.Now()
	date := now.UTC().Format("20060102T150405Z")
	authDate := date[:8]
	request.Header.Set("X-Date", date)

	payload := hex.EncodeToString(hashSHA256(body))
	request.Header.Set("X-Content-Sha256", payload)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	queryString := strings.Replace(queries.Encode(), "+", "%20", -1)
	signedHeaders := []string{"host", "x-date", "x-content-sha256", "content-type"}
	var headerList []string
	for _, header := range signedHeaders {
		if header == "host" {
			headerList = append(headerList, header+":"+request.Host)
		} else {
			v := request.Header.Get(header)
			headerList = append(headerList, header+":"+strings.TrimSpace(v))
		}
	}
	headerString := strings.Join(headerList, "\n")
	canonicalString := strings.Join([]string{
		v.method,
		v.path,
		queryString,
		headerString + "\n",
		strings.Join(signedHeaders, ";"),
		payload,
	}, "\n")
	//log.Printf("canonical string:\n%s\n", canonicalString)

	hashedCanonicalString := hex.EncodeToString(hashSHA256([]byte(canonicalString)))
	//log.Printf("hashed canonical string: %s\n", hashedCanonicalString)

	credentialScope := authDate + "/" + v.region + "/" + v.service + "/request"
	signString := strings.Join([]string{
		"HMAC-SHA256",
		date,
		credentialScope,
		hashedCanonicalString,
	}, "\n")
	//log.Printf("sign string:\n%s\n", signString)

	// 3. 构建认证请求头
	signedKey := getSignedKey(v.secretKey, authDate, v.region, v.service)
	signature := hex.EncodeToString(hmacSHA256(signedKey, signString))
	///log.Printf("signature: %s\n", signature)

	authorization := "HMAC-SHA256" +
		" Credential=" + v.accessKey + "/" + credentialScope +
		", SignedHeaders=" + strings.Join(signedHeaders, ";") +
		", Signature=" + signature
	request.Header.Set("Authorization", authorization)
	//log.Printf("authorization: %s\n", authorization)

	// 4. 打印请求，发起请求
	requestRaw, err := httputil.DumpRequest(request, true)
	if err != nil {
		return nil, fmt.Errorf("dump request err: %w", err)
	}

	log.Printf("request:\n%s\n", string(requestRaw))

	return request, nil
}
