package storage

import (
	"context"
	"net/http"

	"ewintr.nl/yogai/model"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/fault"
	"github.com/weaviate/weaviate/entities/models"
)

const (
	className = "Video"
)

type Weaviate struct {
	client *weaviate.Client
}

func NewWeaviate(host, weaviateApiKey, openaiApiKey string) (*Weaviate, error) {
	config := weaviate.Config{
		Scheme:     "https",
		Host:       host,
		AuthConfig: auth.ApiKey{Value: weaviateApiKey},
		Headers: map[string]string{
			"X-OpenAI-Api-Key": openaiApiKey,
		},
	}

	c, err := weaviate.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Weaviate{client: c}, nil
}

func (w *Weaviate) ResetSchema() error {

	// delete old
	if err := w.client.Schema().ClassDeleter().WithClassName(className).Do(context.Background()); err != nil {
		// Weaviate will return a 400 if the class does not exist, so this is allowed, only return an error if it's not a 400
		if status, ok := err.(*fault.WeaviateClientError); ok && status.StatusCode != http.StatusBadRequest {
			return err
		}
	}

	// create new
	classObj := &models.Class{
		Class:      className,
		Vectorizer: "text2vec-openai",
		ModuleConfig: map[string]any{
			"text2vec-openai": map[string]any{
				"model":        "ada",
				"modelVersion": "002",
				"type":         "text",
			},
		},
	}

	return w.client.Schema().ClassCreator().WithClass(classObj).Do(context.Background())
}

func (w *Weaviate) Save(ctx context.Context, video *model.Video) error {
	vID := video.ID.String()
	// check it already exists
	exists, err := w.client.Data().
		Checker().
		WithID(vID).
		WithClassName(className).
		Do(ctx)
	if err != nil {
		return err
	}

	if exists {
		return w.client.Data().
			Updater().
			WithID(vID).
			WithClassName(className).
			WithProperties(video).
			Do(ctx)
	}

	_, err = w.client.Data().
		Creator().
		WithClassName(className).
		WithID(vID).
		WithProperties(video).
		Do(ctx)

	return err
}
