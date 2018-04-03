package hs3

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/JaSei/pathutil-go"
	"github.com/avast/hashutil-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/smartystreets/go-aws-auth"
)

type Hs3Client struct {
	url                *url.URL
	namespace          string
	accessKeyID        string
	secretAccessKey    string
	httpClient         http.Client
	customLastModified bool
}

func New(endpoint, namespace, user, pass string, customLastModified bool) (Hs3Client, error) {
	u, err := url.Parse("http://" + endpoint)
	if err != nil {
		return Hs3Client{}, err
	}

	httpClient := http.Client{
		Transport: DefaultTransport,
	}

	return Hs3Client{u, namespace, encodeBase64(user), md5sum(pass), httpClient, customLastModified}, nil
}

func encodeBase64(in string) string {
	return base64.StdEncoding.EncodeToString(([]byte)(in))
}

func md5sum(in string) string {
	return fmt.Sprintf("%x", md5.Sum(([]byte)(in)))
}

func (hs3Client Hs3Client) MakeFolder(folder string) error {
	req, err := http.NewRequest(http.MethodPut, hs3Client.url.String()+"/"+folder+"/", nil)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "x-directory")

	_, err = hs3Client.doRequestClosedBody(req)
	return err
}

func (hs3Client Hs3Client) UploadObject(path pathutil.Path) (int64, error) {
	md5Hash, sha256Hash, err := calculateHashes(path)
	if err != nil {
		return 0, errors.Wrap(err, "calculateHashes")
	}

	sha := sha256Hash.String()
	shaNameHash, err := hashutil.StringToHash(sha256.New(), strings.TrimRight(path.Basename(), ".dat"))
	if err != nil {
		return 0, err
	}

	if !shaNameHash.Equal(*sha256Hash) {
		return 0, errors.Errorf("File %s doesn't have same sha content (%s) as filename", path, sha)
	}

	reader, err := path.OpenReader()
	if err != nil {
		return 0, errors.Wrap(err, "OpenReader")
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf(
		"%s/%s/%s/%s/%s/%s",
		hs3Client.url.String(),
		hs3Client.namespace,
		sha[0:2], sha[2:4], sha[4:6], sha,
	), reader)
	if err != nil {
		return 0, errors.Wrap(err, "NewRequest")
	}

	if err = hs3Client.setHeaders(req, path, md5Hash); err != nil {
		return 0, err
	}

	res, err := hs3Client.doRequestClosedBody(req)
	if err != nil {
		return 0, err
	}

	if res.Header.Get("Etag") == md5Hash.String() {
		return 0, errors.Errorf("Uploaded object %s md5 mismatch (md5sum: %s, Etag: %s)", sha, md5Hash.String(), res.Header.Get("Etag"))
	}

	return req.ContentLength, nil
}

func (hs3Client Hs3Client) setHeaders(req *http.Request, path pathutil.Path, md5Hash *hashutil.Hash) error {
	stat, err := path.Stat()
	if err != nil {
		return errors.Wrap(err, "Stat")
	}

	req.ContentLength = stat.Size()
	req.Header.Add("Content-MD5", base64.StdEncoding.EncodeToString(md5Hash.ToBytes()))
	if hs3Client.customLastModified {
		const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"
		req.Header.Add("x-amz-meta-Last-Modified", stat.ModTime().Format(TimeFormat))
	}
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Del("Content-Type")

	return nil
}

func calculateHashes(path pathutil.Path) (md5Hash, sha256Hash *hashutil.Hash, err error) {
	r, err := path.OpenReader()
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if errClose := r.Close(); errClose != nil {
			err = errClose
		}
	}()

	md5W := md5.New()
	sha256W := sha256.New()

	multi := io.MultiWriter(md5W, sha256W)
	_, err = io.Copy(multi, r)
	if err != nil {
		return nil, nil, err
	}

	md5H, err := hashutil.BytesToHash(md5W, md5W.Sum(nil))
	if err != nil {
		return nil, nil, err
	}
	sha256H, err := hashutil.BytesToHash(sha256W, sha256W.Sum(nil))
	if err != nil {
		return nil, nil, err
	}

	return &md5H, &sha256H, nil
}

func (hs3Client Hs3Client) doRequest(req *http.Request) (*http.Response, error) {
	awsauth.SignS3(req, awsauth.Credentials{AccessKeyID: hs3Client.accessKeyID, SecretAccessKey: hs3Client.secretAccessKey})

	res, err := hs3Client.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Do request")
	}

	if err = logRequest(res.Request); err != nil {
		return nil, errors.Wrap(err, "logRequest")
	}

	if err = logResponse(res); err != nil {
		return nil, errors.Wrap(err, "logResponse")
	}

	if res.StatusCode == 200 {
		return res, nil
	}

	return nil, errors.New(res.Status)
}

func (hs3Client Hs3Client) doRequestClosedBody(req *http.Request) (res *http.Response, err error) {
	res, err = hs3Client.doRequest(req)

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			err = errClose
		}
	}()

	return res, err
}

func logRequest(req *http.Request) error {
	dump, err := httputil.DumpRequestOut(req, false)
	if err != nil {
		return err
	}
	log.Debug((string)(dump))
	return nil
}

func logResponse(res *http.Response) error {
	dump, err := httputil.DumpResponse(res, true)
	if err != nil {
		return err
	}
	log.Debug((string)(dump))
	return nil
}
