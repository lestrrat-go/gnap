package client

import (
	"net/http"

	"github.com/lestrrat-go/option"
)

type identHTTPClient struct{}

type ClientOption interface {
	option.Interface
	clientOption()
}

type clientOption struct {
	option.Interface
}

func (*clientOption) clientOption() {}

func WithHTTPClient(v *http.Client) ClientOption {
	return &clientOption{
		option.New(identHTTPClient{}, v),
	}
}
