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

type ClientDisplay struct {
	logo_uri    *string
	name        *string
	uri         *string
	extraFields map[string]interface{}
}

func (c *ClientDisplay) Get(key string) (interface{}, bool) {
	switch key {
	case "logo_uri":
		if c.logo_uri == nil {
			return nil, false
		}
		return c.logo_uri, true
	case "name":
		if c.name == nil {
			return nil, false
		}
		return c.name, true
	case "uri":
		if c.uri == nil {
			return nil, false
		}
		return c.uri, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *ClientDisplay) Set(key string, value interface{}) error {
	switch key {
	case "logo_uri":
		if v, ok := value.(string); ok {
			c.logo_uri = &v
		} else if value == nil {
			c.logo_uri = nil
		} else {
			return errors.Errorf(`invalid type for "logo_uri" (%T)`, value)
		}
	case "name":
		if v, ok := value.(string); ok {
			c.name = &v
		} else if value == nil {
			c.name = nil
		} else {
			return errors.Errorf(`invalid type for "name" (%T)`, value)
		}
	case "uri":
		if v, ok := value.(string); ok {
			c.uri = &v
		} else if value == nil {
			c.uri = nil
		} else {
			return errors.Errorf(`invalid type for "uri" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *ClientDisplay) SetLogoURI(v string) {
	c.logo_uri = &v
}

func (c *ClientDisplay) LogoURI() string {
	if c.logo_uri == nil {
		return ""
	}
	return *(c.logo_uri)
}

func (c *ClientDisplay) SetName(v string) {
	c.name = &v
}

func (c *ClientDisplay) Name() string {
	if c.name == nil {
		return ""
	}
	return *(c.name)
}

func (c *ClientDisplay) SetURI(v string) {
	c.uri = &v
}

func (c *ClientDisplay) URI() string {
	if c.uri == nil {
		return ""
	}
	return *(c.uri)
}

func (c ClientDisplay) MarshalJSON() ([]byte, error) {
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

func (c *ClientDisplay) UnmarshalJSON(data []byte) error {
	c.logo_uri = nil
	c.name = nil
	c.uri = nil
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
			case "logo_uri":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading logo_uri`)
				}
				c.logo_uri = &tmp
			case "name":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading name`)
				}
				c.name = &tmp
			case "uri":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading uri`)
				}
				c.uri = &tmp
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

func (c *ClientDisplay) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.logo_uri; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "logo_uri", Value: *tmp})
	}
	if tmp := c.name; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "name", Value: *tmp})
	}
	if tmp := c.uri; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "uri", Value: *tmp})
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

func (c *ClientDisplay) Iterate(ctx context.Context) mapiter.Iterator {
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
