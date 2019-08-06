package worker

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"	
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/whosonfirst/go-whosonfirst-aws/session"
	"github.com/go-spatial/geom/slippy"	
	"strings"
)

type SQSWorker struct {
	Worker
	client    *sqs.SQS
	queue_url string
}

func NewSQSWorker(dsn map[string]string, queue_url string) (Worker, error) {

	creds, ok := dsn["credentials"]

	if !ok {
		return nil, errors.New("Missing credentials")
	}

	region, ok := dsn["region"]

	if !ok {
		return nil, errors.New("Missing region")
	}

	sess, err := session.NewSessionWithCredentials(creds, region)

	if err != nil {
		return nil, err
	}

	client := sqs.New(sess)

	if !strings.HasPrefix(queue_url, "https://sqs") {

		rsp, err := client.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: aws.String(queue_url),
		})

		if err != nil {
			return nil, err
		}

		queue_url = *rsp.QueueUrl
	}

	w := SQSWorker{
		client:    client,
		queue_url: queue_url,
	}

	return &w, nil
}

func (w *SQSWorker) RenderRasterzenTile(t slippy.Tile) error {
	return w.renderTile(t, "rasterzen", "json")
}

func (w *SQSWorker) RenderGeoJSONTile(t slippy.Tile) error {
	return w.renderTile(t, "geojson", "geojson")
}

func (w *SQSWorker) RenderExtentTile(t slippy.Tile) error {
	return w.renderTile(t, "extent", "geojson")
}

func (w *SQSWorker) RenderSVGTile(t slippy.Tile) error {
	return w.renderTile(t, "svg", "svg")
}

func (w *SQSWorker) RenderPNGTile(t slippy.Tile) error {
	return w.renderTile(t, "png", "png")
}

func (w *SQSWorker) renderTile(t slippy.Tile, prefix, format string) error {

	body := "fix me"

	msg := &sqs.SendMessageInput{
		QueueUrl:    aws.String(w.queue_url),
		MessageBody: aws.String(body),
	}

	_, err := w.client.SendMessage(msg)
	return err
}