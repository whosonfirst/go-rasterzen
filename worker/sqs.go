package worker

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-spatial/geom/slippy"
	"github.com/whosonfirst/go-whosonfirst-aws/session"
	"strings"
)

// TO DO: store/pass Nextzen options here...

type SQSMessage struct {
	Z      uint   `json:"z"`
	X      uint   `json:"x"`
	Y      uint   `json:"y"`
	Prefix string `json:"prefix"`
}

type SQSWorker struct {
	Worker
	client    *sqs.SQS
	queue_url string
}

func NewSQSWorker(dsn map[string]string) (Worker, error) {

	creds, ok := dsn["credentials"]

	if !ok {
		return nil, errors.New("Missing credentials")
	}

	region, ok := dsn["region"]

	if !ok {
		return nil, errors.New("Missing region")
	}

	queue_url, ok := dsn["queue"]

	if !ok {
		return nil, errors.New("Missing queue")
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

func (w *SQSWorker) renderTile(t slippy.Tile, prefix string, format string) error {

	// see notes above
	
	msg := SQSMessage{
		Z:      t.Z,
		X:      t.X,
		Y:      t.Y,
		Prefix: prefix,
	}

	enc, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	req := &sqs.SendMessageInput{
		QueueUrl:    aws.String(w.queue_url),
		MessageBody: aws.String(string(enc)),
	}

	_, err = w.client.SendMessage(req)
	return err
}
