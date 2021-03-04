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

type SubjectRequest struct {
	assertions  []string
	subIDs      []string
	extraFields map[string]interface{}
}

func (c *SubjectRequest) Get(key string) (interface{}, bool) {
	switch key {
	case "assertions":
		if len(c.assertions) == 0 {
			return nil, false
		}
		return c.assertions, true
	case "sub_i_ds":
		if len(c.subIDs) == 0 {
			return nil, false
		}
		return c.subIDs, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *SubjectRequest) Set(key string, value interface{}) error {
	switch key {
	case "assertions":
		if v, ok := value.([]string); ok {
			c.assertions = v
		} else {
			return errors.Errorf(`invalid type for "assertions" (%T)`, value)
		}
	case "sub_i_ds":
		if v, ok := value.([]string); ok {
			c.subIDs = v
		} else {
			return errors.Errorf(`invalid type for "sub_i_ds" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *SubjectRequest) AddAssertions(v ...string) {
	c.assertions = append(c.assertions, v...)
}

func (c *SubjectRequest) AddSubIDs(v ...string) {
	c.subIDs = append(c.subIDs, v...)
}

func (c SubjectRequest) MarshalJSON() ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	buf.WriteByte('{')
	var i int
	for iter := c.Iterate(ctx); iter.Next(ctx); {
		if i > 0 {
			buf.WriteByte(',')
		}
		i++
		pair := iter.Pair()
		buf.WriteString(strconv.Quote(pair.Key.(string)))
		buf.WriteByte(':')
		if err := enc.Encode(pair.Value); err != nil {
			return nil, errors.Wrapf(err, `failed to encode %s`, pair.Key.(string))
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (c *SubjectRequest) UnmarshalJSON(data []byte) error {
	c.assertions = nil
	c.subIDs = nil
	dec := json.NewDecoder(bytes.NewReader(data))
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
			} else if tok != '{' {
				return errors.Errorf(`expected '{', but got '%c'`, tok)
			}
		case string:
			switch tok {
			case "assertions":
				if err := dec.Decode(&(c.assertions)); err != nil {
					return errors.Wrap(err, `error reading assertions`)
				}
			case "sub_i_ds":
				if err := dec.Decode(&(c.subIDs)); err != nil {
					return errors.Wrap(err, `error reading sub_i_ds`)
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

func (c *SubjectRequest) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.assertions; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "assertions", Value: tmp})
	}
	if tmp := c.subIDs; len(tmp) > 0 {
		pairs = append(pairs, &mapiter.Pair{Key: "sub_i_ds", Value: tmp})
	}
	var extraKeys []string
	for k := range c.extraFields {
		extraKeys = append(extraKeys, k)
	}
	sort.Strings(extraKeys)
	for _, k := range extraKeys {
		pairs = append(pairs, &mapiter.Pair{Key: k, Value: c.extraFields[k]})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key.(string) < pairs[j].Key.(string)
	})
	return pairs
}

func (c *SubjectRequest) Iterate(ctx context.Context) mapiter.Iterator {
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
