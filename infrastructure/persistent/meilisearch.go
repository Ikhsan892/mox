package persistent

import (
	"mox/pkg/config"
	"github.com/meilisearch/meilisearch-go"
)

func NewMeilisearch(cfg config.Database, meiliCfg meilisearch.ClientConfig) *meilisearch.Client {
	return meilisearch.NewClient(meiliCfg)
}
