package client

import (
	"net/http"
)

type Client struct {
	httpcl *http.Client
}

func New(options ...ClientOption) *Client {
	httpcl := http.DefaultClient
	for _, option := range options {
		switch option.Ident() {
		case identHTTPClient{}:
			httpcl = option.Value().(*http.Client)
		}
	}

	return &Client{
		httpcl: httpcl,
	}
}
