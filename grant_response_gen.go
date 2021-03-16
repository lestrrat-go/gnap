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

type GrantResponse struct {
	accessToken  *AccessToken
	continuation *RequestContinuation
	error        *string
	interact     *InteractionResponse
	extraFields  map[string]interface{}
}

func NewGrantResponse() *GrantResponse {
	return &GrantResponse{}
}

func (c *GrantResponse) Validate() error {
	return nil
}

func (c *GrantResponse) Get(key string) (interface{}, bool) {
	switch key {
	case "access_token":
		if c.accessToken == nil {
			return nil, false
		}
		return c.accessToken, true
	case "continue":
		if c.continuation == nil {
			return nil, false
		}
		return c.continuation, true
	case "error":
		if c.error == nil {
			return nil, false
		}
		return c.error, true
	case "interact":
		if c.interact == nil {
			return nil, false
		}
		return c.interact, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *GrantResponse) Set(key string, value interface{}) error {
	switch key {
	case "access_token":
		if v, ok := value.(*AccessToken); ok {
			c.accessToken = v
		} else {
			return errors.Errorf(`invalid type for "access_token" (%T)`, value)
		}
	case "continue":
		if v, ok := value.(*RequestContinuation); ok {
			c.continuation = v
		} else {
			return errors.Errorf(`invalid type for "continue" (%T)`, value)
		}
	case "error":
		if v, ok := value.(string); ok {
			c.error = &v
		} else if value == nil {
			c.error = nil
		} else {
			return errors.Errorf(`invalid type for "error" (%T)`, value)
		}
	case "interact":
		if v, ok := value.(*InteractionResponse); ok {
			c.interact = v
		} else {
			return errors.Errorf(`invalid type for "interact" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *GrantResponse) SetAccessToken(v *AccessToken) {
	c.accessToken = v
}

func (c *GrantResponse) AccessToken() *AccessToken {
	return c.accessToken
}

func (c *GrantResponse) SetContinue(v *RequestContinuation) {
	c.continuation = v
}

func (c *GrantResponse) Continue() *RequestContinuation {
	return c.continuation
}

func (c *GrantResponse) SetError(v string) {
	c.error = &v
}

func (c *GrantResponse) Error() string {
	if c.error == nil {
		return ""
	}
	return *(c.error)
}

func (c *GrantResponse) SetInteract(v *InteractionResponse) {
	c.interact = v
}

func (c *GrantResponse) Interact() *InteractionResponse {
	return c.interact
}

func (c GrantResponse) MarshalJSON() ([]byte, error) {
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
		if err := enc.Encode(pair.Value); err != nil {
			return nil, errors.Wrapf(err, `failed to encode %s`, pair.Key.(string))
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (c *GrantResponse) UnmarshalJSON(data []byte) error {
	c.accessToken = nil
	c.continuation = nil
	c.error = nil
	c.interact = nil
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
				if err := dec.Decode(&(c.accessToken)); err != nil {
					return errors.Wrap(err, `error reading access_token`)
				}
			case "continue":
				if err := dec.Decode(&(c.continuation)); err != nil {
					return errors.Wrap(err, `error reading continue`)
				}
			case "error":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading error`)
				}
				c.error = &tmp
			case "interact":
				if err := dec.Decode(&(c.interact)); err != nil {
					return errors.Wrap(err, `error reading interact`)
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

func (c *GrantResponse) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.accessToken; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "access_token", Value: *tmp})
	}
	if tmp := c.continuation; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "continue", Value: *tmp})
	}
	if tmp := c.error; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "error", Value: *tmp})
	}
	if tmp := c.interact; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "interact", Value: *tmp})
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

func (c *GrantResponse) Iterate(ctx context.Context) mapiter.Iterator {
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
