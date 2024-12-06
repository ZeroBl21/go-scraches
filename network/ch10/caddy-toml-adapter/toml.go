package tomladapter

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/pelletier/go-toml"
)

func init() {
	caddyconfig.RegisterAdapter("toml", Adapter{})
}

type Adapter struct{}

func (a Adapter) Adapt(body []byte, _ map[string]any) ([]byte, []caddyconfig.Warning, error) {
	tree, err := toml.LoadBytes(body)
	if err != nil {
	}

	b, err := json.Marshal(tree.ToMap())

	return b, nil, err
}
