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

type InteractionResponse struct {
	app         *string
	finish      *string
	redirect    *string
	userCode    *string
	extraFields map[string]interface{}
}

func NewInteractionResponse() *InteractionResponse {
	return &InteractionResponse{}
}

func (c *InteractionResponse) Validate() error {
	return nil
}

func (c *InteractionResponse) Get(key string) (interface{}, bool) {
	switch key {
	case "app":
		if c.app == nil {
			return nil, false
		}
		return c.app, true
	case "finish":
		if c.finish == nil {
			return nil, false
		}
		return c.finish, true
	case "redirect":
		if c.redirect == nil {
			return nil, false
		}
		return c.redirect, true
	case "user_code":
		if c.userCode == nil {
			return nil, false
		}
		return c.userCode, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *InteractionResponse) Set(key string, value interface{}) error {
	switch key {
	case "app":
		if v, ok := value.(string); ok {
			c.app = &v
		} else if value == nil {
			c.app = nil
		} else {
			return errors.Errorf(`invalid type for "app" (%T)`, value)
		}
	case "finish":
		if v, ok := value.(string); ok {
			c.finish = &v
		} else if value == nil {
			c.finish = nil
		} else {
			return errors.Errorf(`invalid type for "finish" (%T)`, value)
		}
	case "redirect":
		if v, ok := value.(string); ok {
			c.redirect = &v
		} else if value == nil {
			c.redirect = nil
		} else {
			return errors.Errorf(`invalid type for "redirect" (%T)`, value)
		}
	case "user_code":
		if v, ok := value.(string); ok {
			c.userCode = &v
		} else if value == nil {
			c.userCode = nil
		} else {
			return errors.Errorf(`invalid type for "user_code" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *InteractionResponse) SetApp(v string) {
	c.app = &v
}

func (c *InteractionResponse) App() string {
	if c.app == nil {
		return ""
	}
	return *(c.app)
}

func (c *InteractionResponse) SetFinish(v string) {
	c.finish = &v
}

func (c *InteractionResponse) Finish() string {
	if c.finish == nil {
		return ""
	}
	return *(c.finish)
}

func (c *InteractionResponse) SetRedirect(v string) {
	c.redirect = &v
}

func (c *InteractionResponse) Redirect() string {
	if c.redirect == nil {
		return ""
	}
	return *(c.redirect)
}

func (c *InteractionResponse) SetUserCode(v string) {
	c.userCode = &v
}

func (c *InteractionResponse) UserCode() string {
	if c.userCode == nil {
		return ""
	}
	return *(c.userCode)
}

func (c InteractionResponse) MarshalJSON() ([]byte, error) {
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

func (c *InteractionResponse) UnmarshalJSON(data []byte) error {
	c.app = nil
	c.finish = nil
	c.redirect = nil
	c.userCode = nil
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
			case "app":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading app`)
				}
				c.app = &tmp
			case "finish":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading finish`)
				}
				c.finish = &tmp
			case "redirect":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading redirect`)
				}
				c.redirect = &tmp
			case "user_code":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading user_code`)
				}
				c.userCode = &tmp
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

func (c *InteractionResponse) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.app; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "app", Value: *tmp})
	}
	if tmp := c.finish; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "finish", Value: *tmp})
	}
	if tmp := c.redirect; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "redirect", Value: *tmp})
	}
	if tmp := c.userCode; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "user_code", Value: *tmp})
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

func (c *InteractionResponse) Iterate(ctx context.Context) mapiter.Iterator {
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
