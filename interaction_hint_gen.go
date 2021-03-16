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

type InteractionHint struct {
	uiLocales   []string
	extraFields map[string]interface{}
}

func NewInteractionHint() *InteractionHint {
	return &InteractionHint{}
}

func (c *InteractionHint) Validate() error {
	return nil
}

func (c *InteractionHint) Get(key string) (interface{}, bool) {
	switch key {
	case "ui_locales":
		if len(c.uiLocales) == 0 {
			return nil, false
		}
		return c.uiLocales, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *InteractionHint) Set(key string, value interface{}) error {
	switch key {
	case "ui_locales":
		if v, ok := value.([]string); ok {
			c.uiLocales = v
		} else {
			return errors.Errorf(`invalid type for "ui_locales" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *InteractionHint) AddUiLocales(v ...string) *InteractionHint {
	c.uiLocales = append(c.uiLocales, v...)
	return c
}

func (c *InteractionHint) UiLocales() []string {
	return c.uiLocales
}

func (c InteractionHint) MarshalJSON() ([]byte, error) {
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

func (c *InteractionHint) UnmarshalJSON(data []byte) error {
	c.uiLocales = nil
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
			case "ui_locales":
				if err := dec.Decode(&(c.uiLocales)); err != nil {
					return errors.Wrap(err, `error reading ui_locales`)
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

func (c *InteractionHint) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.uiLocales; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "ui_locales", Value: tmp})
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

func (c *InteractionHint) Iterate(ctx context.Context) mapiter.Iterator {
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
