package s3

import (
	"bufio"
	"bytes"
	"github.com/aws/aws-sdk-go/aws/awserr"
	aws_s3 "github.com/aws/aws-sdk-go/service/s3"
	wof_s3 "github.com/whosonfirst/go-whosonfirst-aws/s3"
	"github.com/whosonfirst/go-whosonfirst-cache"
	"io"
	"io/ioutil"
)

type S3Cache struct {
	cache.Cache
	conn *wof_s3.S3Connection
}

func NewS3Cache(dsn string) (cache.Cache, error) {

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
