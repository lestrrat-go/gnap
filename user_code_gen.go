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

type UserCode struct {
	code        *string
	url         *string
	extraFields map[string]interface{}
}

func NewUserCode(code string) *UserCode {
	return &UserCode{
		code: &code,
	}
}

func (c *UserCode) Validate() error {
	if c.code == nil {
		return errors.Errorf(`field "code" is required`)
	}
	return nil
}

func (c *UserCode) Get(key string) (interface{}, bool) {
	switch key {
	case "code":
		if c.code == nil {
			return nil, false
		}
		return c.code, true
	case "url":
		if c.url == nil {
			return nil, false
		}
		return c.url, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *UserCode) Set(key string, value interface{}) error {
	switch key {
	case "code":
		if v, ok := value.(string); ok {
			c.code = &v
		} else if value == nil {
			c.code = nil
		} else {
			return errors.Errorf(`invalid type for "code" (%T)`, value)
		}
	case "url":
		if v, ok := value.(string); ok {
			c.url = &v
		} else if value == nil {
			c.url = nil
		} else {
			return errors.Errorf(`invalid type for "url" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *UserCode) SetCode(v string) {
	c.code = &v
}

func (c *UserCode) Code() string {
	if c.code == nil {
		return ""
	}
	return *(c.code)
}

func (c *UserCode) SetURL(v string) {
	c.url = &v
}

func (c *UserCode) URL() string {
	if c.url == nil {
		return ""
	}
	return *(c.url)
}

func (c UserCode) MarshalJSON() ([]byte, error) {
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

func (c *UserCode) UnmarshalJSON(data []byte) error {
	c.code = nil
	c.url = nil
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
			case "code":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading code`)
				}
				c.code = &tmp
			case "url":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading url`)
				}
				c.url = &tmp
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

func (c *UserCode) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.code; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "code", Value: *tmp})
	}
	if tmp := c.url; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "url", Value: *tmp})
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

func (c *UserCode) Iterate(ctx context.Context) mapiter.Iterator {
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
