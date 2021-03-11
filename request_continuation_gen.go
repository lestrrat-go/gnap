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

type RequestContinuation struct {
	accessToken *AccessToken
	uri         *string
	wait        *int64
	extraFields map[string]interface{}
}

func NewRequestContinuation(accessToken AccessToken, uri string) *RequestContinuation {
	return &RequestContinuation{
		accessToken: &accessToken,
		uri:         &uri,
	}
}

func (c *RequestContinuation) Validate() error {
	if c.accessToken == nil {
		return errors.Errorf(`field "accessToken" is required`)
	}
	if c.uri == nil {
		return errors.Errorf(`field "uri" is required`)
	}
	return nil
}

func (c *RequestContinuation) Get(key string) (interface{}, bool) {
	switch key {
	case "access_token":
		if c.accessToken == nil {
			return nil, false
		}
		return c.accessToken, true
	case "uri":
		if c.uri == nil {
			return nil, false
		}
		return c.uri, true
	case "wait":
		if c.wait == nil {
			return nil, false
		}
		return c.wait, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *RequestContinuation) Set(key string, value interface{}) error {
	switch key {
	case "access_token":
		if v, ok := value.(*AccessToken); ok {
			c.accessToken = v
		} else {
			return errors.Errorf(`invalid type for "access_token" (%T)`, value)
		}
	case "uri":
		if v, ok := value.(string); ok {
			c.uri = &v
		} else if value == nil {
			c.uri = nil
		} else {
			return errors.Errorf(`invalid type for "uri" (%T)`, value)
		}
	case "wait":
		if v, ok := value.(*int64); ok {
			c.wait = v
		} else {
			return errors.Errorf(`invalid type for "wait" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *RequestContinuation) SetAccessToken(v *AccessToken) {
	c.accessToken = v
}

func (c *RequestContinuation) AccessToken() *AccessToken {
	return c.accessToken
}

func (c *RequestContinuation) SetURI(v string) {
	c.uri = &v
}

func (c *RequestContinuation) URI() string {
	if c.uri == nil {
		return ""
	}
	return *(c.uri)
}

func (c *RequestContinuation) SetWait(v *int64) {
	c.wait = v
}

func (c *RequestContinuation) Wait() *int64 {
	return c.wait
}

func (c RequestContinuation) MarshalJSON() ([]byte, error) {
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

func (c *RequestContinuation) UnmarshalJSON(data []byte) error {
	c.accessToken = nil
	c.uri = nil
	c.wait = nil
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
			case "uri":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading uri`)
				}
				c.uri = &tmp
			case "wait":
				if err := dec.Decode(&(c.wait)); err != nil {
					return errors.Wrap(err, `error reading wait`)
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

func (c *RequestContinuation) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.accessToken; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "access_token", Value: *tmp})
	}
	if tmp := c.uri; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "uri", Value: *tmp})
	}
	if tmp := c.wait; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "wait", Value: *tmp})
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

func (c *RequestContinuation) Iterate(ctx context.Context) mapiter.Iterator {
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
