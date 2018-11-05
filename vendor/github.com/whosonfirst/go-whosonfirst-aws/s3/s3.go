package s3

// ideally we could make this conform to some standard "storage" interface
// but that works hasn't been done (20180120/thisisaaronland)

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/whosonfirst/go-whosonfirst-mimetypes"
	"io"
	_ "log"
	"os/user"
	"path/filepath"
	"strings"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func ReadCloserFromBytes(b []byte) (io.ReadCloser, error) {
	body := bytes.NewReader(b)
	return nopCloser{body}, nil
}

type S3Connection struct {
	session *session.Session
	service *s3.S3
	bucket  string
	prefix  string
}

type S3Config struct {
	Bucket      string
	Prefix      string
	Region      string
	Credentials string // see notes below
}

func ValidS3Credentials() []string {

	valid := []string{
		"env:",
		"iam:",
		"{PROFILE}",
		"{PATH}:{PROFILE}",
	}

	return valid
}

func ValidS3CredentialsString() string {

	valid := ValidS3Credentials()
	return fmt.Sprintf("Valid credential flags are: %s", strings.Join(valid, ", "))
}

func NewS3ConfigFromString(str_config string) (*S3Config, error) {

	config := S3Config{
		Bucket:      "",
		Prefix:      "",
		Region:      "",
		Credentials: "",
	}

	str_config = strings.Trim(str_config, " ")

	if str_config != "" {
		parts := strings.Split(str_config, " ")

		for _, p := range parts {

			p = strings.Trim(p, " ")
			kv := strings.Split(p, "=")

			if len(kv) != 2 {
				return nil, errors.New("Invalid count for config block")
			}

			switch kv[0] {
			case "bucket":
				config.Bucket = kv[1]
			case "prefix":
				config.Prefix = kv[1]
			case "region":
				config.Region = kv[1]
			case "credentials":
				config.Credentials = kv[1]
			default:
				return nil, errors.New("Invalid key for config block")
			}
		}
	}

	if config.Bucket == "" {
		return nil, errors.New("Missing bucket config")
	}

	if config.Region == "" {
		return nil, errors.New("Missing region config")
	}

	if config.Credentials == "" {
		return nil, errors.New("Missing credentials config")
	}

	return &config, nil
}

func NewS3Connection(s3cfg *S3Config) (*S3Connection, error) {

	if s3cfg.Bucket == "" {
		return nil, errors.New("Invalid S3 bucket name")
	}

	// https://docs.aws.amazon.com/sdk-for-go/v1/developerguide/configuring-sdk.html
	// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/

	cfg := aws.NewConfig()
	cfg.WithRegion(s3cfg.Region)

	if strings.HasPrefix(s3cfg.Credentials, "env:") {

		creds := credentials.NewEnvCredentials()
		cfg.WithCredentials(creds)

	} else if strings.HasPrefix(s3cfg.Credentials, "iam:") {

		// assume an IAM role suffient for doing whatever

	} else if s3cfg.Credentials != "" {

		details := strings.Split(s3cfg.Credentials, ":")

		var creds_file string
		var profile string

		if len(details) == 1 {

			whoami, err := user.Current()

			if err != nil {
				return nil, err
			}

			dotaws := filepath.Join(whoami.HomeDir, ".aws")
			creds_file = filepath.Join(dotaws, "credentials")

			profile = details[0]

		} else {

			path, err := filepath.Abs(details[0])

			if err != nil {
				return nil, err
			}

			creds_file = path
			profile = details[1]
		}

		creds := credentials.NewSharedCredentials(creds_file, profile)
		cfg.WithCredentials(creds)

	} else {

		// for backwards compatibility as of 05a6042dc5956c13513bdc5ab4969877013f795c
		// (20161203/thisisaaronland)

		creds := credentials.NewEnvCredentials()
		cfg.WithCredentials(creds)
	}

	sess := session.New(cfg)

	if s3cfg.Credentials != "" {

		_, err := sess.Config.Credentials.Get()

		if err != nil {
			return nil, err
		}
	}

	service := s3.New(sess)

	c := S3Connection{
		session: sess,
		service: service,
		bucket:  s3cfg.Bucket,
		prefix:  s3cfg.Prefix,
	}

	return &c, nil
}

func (conn *S3Connection) URI(key string) string {

	key = conn.prepareKey(key)

	if conn.prefix != "" {
		key = fmt.Sprintf("%s/%s", conn.prefix, key)
	}

	return fmt.Sprintf("https://s3.amazonaws.com/%s/%s", conn.bucket, key)
}

// https://tools.ietf.org/html/rfc7231#section-4.3.2
// https://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectHEAD.html

func (conn *S3Connection) Head(key string) (*s3.HeadObjectOutput, error) {

	key = conn.prepareKey(key)

	params := &s3.HeadObjectInput{
		Bucket: aws.String(conn.bucket),
		Key:    aws.String(key),
	}

	rsp, err := conn.service.HeadObject(params)

	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func (conn *S3Connection) Get(key string) (io.ReadCloser, error) {

	key = conn.prepareKey(key)

	params := &s3.GetObjectInput{
		Bucket: aws.String(conn.bucket),
		Key:    aws.String(key),
	}

	rsp, err := conn.service.GetObject(params)

	if err != nil {
		return nil, err
	}

	return rsp.Body, nil
}

func (conn *S3Connection) GetBytes(key string) ([]byte, error) {

	fh, err := conn.Get(key)

	if err != nil {
		return nil, err
	}

	defer fh.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(fh)

	return buf.Bytes(), nil
}

func (conn *S3Connection) Put(key string, fh io.ReadCloser, args ...interface{}) error {

	// file under known knowns: AWS expects a ReadSeeker for performance
	// and memory reasons but we're passing around ReadClosers  - see also:
	// https://github.com/whosonfirst/go-whosonfirst-readwrite/issues/2
	// (20180120/thisisaaronland)

	defer fh.Close()

	parsed := strings.Split(key, "#")

	key = parsed[0]
	key = conn.prepareKey(key)

	uploader := s3manager.NewUploader(conn.session)

	// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/s3manager/#UploadInput

	params := s3manager.UploadInput{
		Bucket: aws.String(conn.bucket),
		Key:    aws.String(key),
		Body:   fh,
	}

	ext := filepath.Ext(key)
	types := mimetypes.TypesByExtension(ext)

	if len(types) == 1 {
		params.ContentType = aws.String(types[0])
	}

	// I don't love this... still working it out
	// (20180120/thisisaaronland)

	if len(parsed) > 1 {

		extras := strings.Split(parsed[1], ",")

		for _, ex := range extras {

			kv := strings.Split(ex, "=")

			if len(kv) != 2 {
				return errors.New("Invalid extras")
			}

			k := kv[0]
			v := kv[1]

			switch k {
			case "ACL":
				params.ACL = aws.String(v)
			case "ContentType":
				params.ContentType = aws.String(v)
			default:
				// pass
			}
		}
	}

	_, err := uploader.Upload(&params)
	return err
}

func (conn *S3Connection) PutBytes(key string, body []byte) error {

	fh, err := ReadCloserFromBytes(body)

	if err != nil {
		return err
	}

	return conn.Put(key, fh)
}

func (conn *S3Connection) Delete(key string) error {

	key = conn.prepareKey(key)

	params := &s3.DeleteObjectInput{
		Bucket: aws.String(conn.bucket),
		Key:    aws.String(key),
	}

	_, err := conn.service.DeleteObject(params)

	if err != nil {
		return err
	}

	return nil
}

func (conn *S3Connection) HasChanged(key string, local []byte) (bool, error) {

	// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#HeadObjectInput
	// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#HeadObjectOutput

	head, err := conn.Head(key)

	if err != nil {

		aws_err := err.(awserr.Error)

		if aws_err.Code() == "NotFound" {
			return true, nil
		}

		if aws_err.Code() == "SlowDown" {

		}

		return false, err
	}

	enc := md5.Sum(local)
	local_hash := hex.EncodeToString(enc[:])

	etag := *head.ETag
	remote_hash := strings.Replace(etag, "\"", "", -1)

	if local_hash == remote_hash {
		return false, nil
	}

	return true, nil
}

func (conn *S3Connection) prepareKey(key string) string {

	if conn.prefix == "" {
		return key
	}

	return filepath.Join(conn.prefix, key)
}
