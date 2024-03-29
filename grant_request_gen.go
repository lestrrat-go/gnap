package gnap

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/lestrrat-go/iter/mapiter"
	"github.com/pkg/errors"
)

type GrantRequest struct {
	accessTokens []*AccessTokenRequest
	capabilities []string
	client       *Client
	interact     *InteractionRequest
	subject      *SubjectRequest
	extraFields  map[string]interface{}
}

func NewGrantRequest() *GrantRequest {
	return &GrantRequest{}
}

func (c *GrantRequest) Validate() error {
	if len(c.accessTokens) > 0 {
		for _, access := range c.accessTokens {
			if access.label == nil {
				return errors.Errorf(`"label" is required in "access" field for multiple access token requests (2.1.1)`)
			}
		}
	}
	return nil
}

func (c *GrantRequest) Get(key string) (interface{}, bool) {
	switch key {
	case "access_token":
		if len(c.accessTokens) == 0 {
			return nil, false
		}
		return c.accessTokens, true
	case "capabilities":
		if len(c.capabilities) == 0 {
			return nil, false
		}
		return c.capabilities, true
	case "client":
		if c.client == nil {
			return nil, false
		}
		return c.client, true
	case "interact":
		if c.interact == nil {
			return nil, false
		}
		return c.interact, true
	case "subject":
		if c.subject == nil {
			return nil, false
		}
		return c.subject, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *GrantRequest) Set(key string, value interface{}) error {
	switch key {
	case "access_token":
		if v, ok := value.([]*AccessTokenRequest); ok {
			c.accessTokens = v
		} else {
			return errors.Errorf(`invalid type for "access_token" (%T)`, value)
		}
	case "capabilities":
		if v, ok := value.([]string); ok {
			c.capabilities = v
		} else {
			return errors.Errorf(`invalid type for "capabilities" (%T)`, value)
		}
	case "client":
		if v, ok := value.(*Client); ok {
			c.client = v
		} else {
			return errors.Errorf(`invalid type for "client" (%T)`, value)
		}
	case "interact":
		if v, ok := value.(*InteractionRequest); ok {
			c.interact = v
		} else {
			return errors.Errorf(`invalid type for "interact" (%T)`, value)
		}
	case "subject":
		if v, ok := value.(*SubjectRequest); ok {
			c.subject = v
		} else {
			return errors.Errorf(`invalid type for "subject" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *GrantRequest) AddAccessTokens(v ...*AccessTokenRequest) *GrantRequest {
	c.accessTokens = append(c.accessTokens, v...)
	return c
}

func (c *GrantRequest) AccessTokens() []*AccessTokenRequest {
	return c.accessTokens
}

func (c *GrantRequest) AddCapabilities(v ...string) *GrantRequest {
	c.capabilities = append(c.capabilities, v...)
	return c
}

func (c *GrantRequest) Capabilities() []string {
	return c.capabilities
}

func (c *GrantRequest) SetClient(v *Client) {
	c.client = v
}

func (c *GrantRequest) Client() *Client {
	return c.client
}

func (c *GrantRequest) SetInteract(v *InteractionRequest) {
	c.interact = v
}

func (c *GrantRequest) Interact() *InteractionRequest {
	return c.interact
}

func (c *GrantRequest) SetSubject(v *SubjectRequest) {
	c.subject = v
}

func (c *GrantRequest) Subject() *SubjectRequest {
	return c.subject
}

func (c GrantRequest) MarshalJSON() ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	buf.WriteByte('{')
	var i int
	for iter := c.Iterate(ctx); iter.Next(ctx); {
		pair := iter.Pair()
		if i > 0 {
			buf.WriteByte(',')
		}
		i++
		buf.WriteString(strconv.Quote(pair.Key.(string)))
		buf.WriteByte(':')
		switch pair.Key.(string) {
		case "access_token":
			v := pair.Value.([]*AccessTokenRequest)
			if len(v) == 1 {
				pair.Value = v[0]
			}
		}
		if err := enc.Encode(pair.Value); err != nil {
			return nil, errors.Wrapf(err, `failed to encode %s`, pair.Key.(string))
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (c *GrantRequest) UnmarshalJSON(data []byte) error {
	c.accessTokens = nil
	c.capabilities = nil
	c.client = nil
	c.interact = nil
	c.subject = nil
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return errors.Wrap(err, `error reading token`)
	}
	switch tok := tok.(type) {
	case json.Delim:
		if tok != '{' {
			return errors.Errorf(`expected '{', but got '%c'`, tok)
		}
	default:
		return errors.Errorf(`expected '{', but got '%c'`, tok)
	}
LOOP:
	for {
		tok, err := dec.Token()
		if err != nil {
			return errors.Wrap(err, `error reading token`)
		}
		switch tok := tok.(type) {
		case json.Delim:
			if tok == '}' { // End of object
				break LOOP
			}
			return errors.Errorf(`unexpected delimiter '%c'`, tok)
		case string:
			switch tok {
			case "access_token":
				var nextThing json.RawMessage
				if err := dec.Decode(&nextThing); err != nil {
					return errors.Wrap(err, `error reading next token access_token`)
				}
				if bytes.HasPrefix(nextThing, []byte{'['}) {
					if err := json.Unmarshal(nextThing, &(c.accessTokens)); err != nil {
						return errors.Wrap(err, `error decoding access_token`)
					}
				} else {
					var tmp AccessTokenRequest
					if err := json.Unmarshal(nextThing, &tmp); err != nil {
						return errors.Wrap(err, `error reading access_token`)
					}
					c.accessTokens = append(c.accessTokens, &tmp)
				}
			case "capabilities":
				if err := dec.Decode(&(c.capabilities)); err != nil {
					return errors.Wrap(err, `error reading capabilities`)
				}
			case "client":
				if err := dec.Decode(&(c.client)); err != nil {
					return errors.Wrap(err, `error reading client`)
				}
			case "interact":
				if err := dec.Decode(&(c.interact)); err != nil {
					return errors.Wrap(err, `error reading interact`)
				}
			case "subject":
				if err := dec.Decode(&(c.subject)); err != nil {
					return errors.Wrap(err, `error reading subject`)
				}
			default:
				var tmp interface{}
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrapf(err, `error reading %s`, tok)
				}
				if c.extraFields == nil {
					c.extraFields = map[string]interface{}{}
				}
				c.extraFields[tok] = tmp
			}
		}
	}
	return nil
}

func (c *GrantRequest) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.accessTokens; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "access_token", Value: tmp})
	}
	if tmp := c.capabilities; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "capabilities", Value: tmp})
	}
	if tmp := c.client; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "client", Value: *tmp})
	}
	if tmp := c.interact; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "interact", Value: *tmp})
	}
	if tmp := c.subject; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "subject", Value: *tmp})
	}
	var extraKeys []string
	for k := range c.extraFields {
		extraKeys = append(extraKeys, k)
	}
	for _, k := range extraKeys {
		pairs = append(pairs, &mapiter.Pair{Key: k, Value: c.extraFields[k]})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key.(string) < pairs[j].Key.(string)
	})
	return pairs
}

func (c *GrantRequest) Iterate(ctx context.Context) mapiter.Iterator {
	pairs := c.makePairs()
	ch := make(chan *mapiter.Pair, len(pairs))
	go func(ctx context.Context, ch chan *mapiter.Pair, pairs []*mapiter.Pair) {
		defer close(ch)
		for _, pair := range pairs {
			select {
			case <-ctx.Done():
				return
			case ch <- pair:
			}
		}
	}(ctx, ch, pairs)
	return mapiter.New(ch)
}
