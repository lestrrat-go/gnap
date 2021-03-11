package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/lestrrat-go/gnap"
	"github.com/pkg/errors"
)

type GrantRequestCmd struct {
	client  *Client
	payload *gnap.GrantRequest
}

func (client *Client) NewGrantRequest() *GrantRequestCmd {
	return &GrantRequestCmd{
		client:  client,
		payload: gnap.NewGrantRequest(),
	}
}

func (cmd *GrantRequestCmd) AddAccessTokens(v ...*gnap.AccessTokenRequest) *GrantRequestCmd {
	cmd.payload.AddAccessTokens(v...)
	return cmd
}

func (cmd *GrantRequestCmd) AddCapabilities(v ...string) *GrantRequestCmd {
	cmd.payload.AddCapabilities(v...)
	return cmd
}

func (cmd *GrantRequestCmd) Client(v *gnap.Client) *GrantRequestCmd {
	cmd.payload.SetClient(v)
	return cmd
}

func (cmd *GrantRequestCmd) Interact(v *gnap.Interaction) *GrantRequestCmd {
	cmd.payload.SetInteract(v)
	return cmd
}

func (cmd *GrantRequestCmd) Subject(v *gnap.SubjectRequest) *GrantRequestCmd {
	cmd.payload.SetSubject(v)
	return cmd
}

func (cmd *GrantRequestCmd) Do(ctx context.Context) error {
	if err := cmd.payload.Validate(); err != nil {
		errors.Wrap(err, `failed to validate payload`)
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(cmd.payload); err != nil {
		return errors.Wrap(err, `failed to encode payload`)
	}
	req, err := http.NewRequest(http.MethodPost, `dummy`, &buf)
	if err != nil {
		return errors.Wrap(err, `failed to create HTTP request`)
	}
	res, err := cmd.client.httpcl.Do(req)
	if err != nil {
		return errors.Wrap(err, `failed to complete HTTP request`)
	}
	_ = res
	return nil
}
