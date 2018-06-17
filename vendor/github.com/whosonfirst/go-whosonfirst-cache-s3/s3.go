package s3

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	aws_s3 "github.com/aws/aws-sdk-go/service/s3"
	wof_s3 "github.com/whosonfirst/go-whosonfirst-aws/s3"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

type S3Cache struct {
	cache.Cache
	conn *wof_s3.S3Connection
	opts *S3CacheOptions
}

type S3CacheOptions struct {
	ACL string
}

func NewS3CacheOptionsFromDefaults() (*S3CacheOptions, error) {

	opts := S3CacheOptions{
		ACL: "private",
	}

	return &opts, nil
}

func NewS3CacheOptionsFromString(str_opts string) (*S3CacheOptions, error) {

	opts, err := NewS3CacheOptionsFromDefaults()

	if err != nil {
		return nil, err
	}

	str_opts = strings.Trim(str_opts, " ")

	if str_opts == "" {
		return opts, nil
	}

	for _, pair := range strings.Split(str_opts, " ") {

		pair = strings.Trim(pair, " ")
		kv := strings.Split(pair, "=")

		if len(kv) != 2 {
			return nil, errors.New("Invalid key=value option")
		}

		switch kv[0] {

		case "ACL":
			opts.ACL = kv[1]
		default:
			return nil, errors.New("Unknown or unsupported key=value pair")
		}
	}

	return opts, nil
}

func NewS3Cache(dsn string, opts *S3CacheOptions) (cache.Cache, error) {

	cfg, err := wof_s3.NewS3ConfigFromString(dsn)

	if err != nil {
		return nil, err
	}

	conn, err := wof_s3.NewS3Connection(cfg)

	if err != nil {
		return nil, err
	}

	c := S3Cache{
		conn: conn,
		opts: opts,
	}

	return &c, nil
}

func (c *S3Cache) Get(key string) (io.ReadCloser, error) {

	fh, err := c.conn.Get(key)

	if err != nil {

		aws_err, ok := err.(awserr.Error)

		if ok {
			switch aws_err.Code() {
			case aws_s3.ErrCodeNoSuchKey:
				err = cache.CacheMiss{}
			default:
				// pass
			}
		}
		return nil, err
	}

	return fh, nil
}

func (c *S3Cache) Set(key string, fh io.ReadCloser) (io.ReadCloser, error) {

	// this is not awesome but until we update all the things (and
	// in particular all the go-whosonfirst-readwrite stuff) to be
	// ReadSeekCloser thingies it's what necessary...
	// (20180617/thisisaaronland)

	var b bytes.Buffer
	wr := bufio.NewWriter(&b)

	io.Copy(wr, fh)
	wr.Flush()

	r := bytes.NewReader(b.Bytes())

	// this sucks, really but the details are still being worked out
	// in go-whosonfirst-aws... (20180617/thisisaaronland)

	key = fmt.Sprintf("%s#ACL=%s", key, c.opts.ACL)
	log.Println("SET", key)

	err := c.conn.Put(key, ioutil.NopCloser(r))

	if err != nil {
		return nil, err
	}

	r.Reset(b.Bytes())
	return ioutil.NopCloser(r), nil
}

func (c *S3Cache) Unset(key string) error {

	return c.conn.Delete(key)
}

func (c *S3Cache) Size() int64 {
	return -1
}

func (c *S3Cache) Hits() int64 {
	return -1
}

func (c *S3Cache) Misses() int64 {
	return -1
}

func (c *S3Cache) Evictions() int64 {
	return -1
}
