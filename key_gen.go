package gnap

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/lestrrat-go/iter/mapiter"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
)

type Key struct {
	cert        *string
	certS256    *string
	jwk         jwk.Key
	proof       *ProofForm
	extraFields map[string]interface{}
}

func (c *Key) Get(key string) (interface{}, bool) {
	switch key {
	case "cert":
		if c.cert == nil {
			return nil, false
		}
		return c.cert, true
	case "cert#S256":
		if c.certS256 == nil {
			return nil, false
		}
		return c.certS256, true
	case "jwk":
		if c.jwk == nil {
			return nil, false
		}
		return c.jwk, true
	case "proof":
		if c.proof == nil {
			return nil, false
		}
		return c.proof, true
	default:
		if c.extraFields == nil {
			return nil, false
		}
		v, ok := c.extraFields[key]
		return v, ok
	}
}

func (c *Key) Set(key string, value interface{}) error {
	switch key {
	case "cert":
		if v, ok := value.(string); ok {
			c.cert = &v
		} else if value == nil {
			c.cert = nil
		} else {
			return errors.Errorf(`invalid type for "cert" (%T)`, value)
		}
	case "cert#S256":
		if v, ok := value.(string); ok {
			c.certS256 = &v
		} else if value == nil {
			c.certS256 = nil
		} else {
			return errors.Errorf(`invalid type for "cert#S256" (%T)`, value)
		}
	case "jwk":
		if v, ok := value.(jwk.Key); ok {
			c.jwk = v
		} else {
			return errors.Errorf(`invalid type for "jwk" (%T)`, value)
		}
	case "proof":
		if v, ok := value.(*ProofForm); ok {
			c.proof = v
		} else {
			return errors.Errorf(`invalid type for "proof" (%T)`, value)
		}
	default:
		if c.extraFields == nil {
			c.extraFields = make(map[string]interface{})
		}
		c.extraFields[key] = value
	}
	return nil
}

func (c *Key) SetCert(v string) {
	c.cert = &v
}

func (c *Key) Cert() string {
	if c.cert == nil {
		return ""
	}
	return *(c.cert)
}

func (c *Key) SetCertS256(v string) {
	c.certS256 = &v
}

func (c *Key) CertS256() string {
	if c.certS256 == nil {
		return ""
	}
	return *(c.certS256)
}

func (c *Key) SetJWK(v jwk.Key) {
	c.jwk = v
}

func (c *Key) JWK() jwk.Key {
	return c.jwk
}

func (c *Key) SetProof(v *ProofForm) {
	c.proof = v
}

func (c *Key) Proof() *ProofForm {
	return c.proof
}

func (c Key) MarshalJSON() ([]byte, error) {
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

func (c *Key) UnmarshalJSON(data []byte) error {
	c.cert = nil
	c.certS256 = nil
	c.jwk = nil
	c.proof = nil
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
			case "cert":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading cert`)
				}
				c.cert = &tmp
			case "cert#S256":
				var tmp string
				if err := dec.Decode(&tmp); err != nil {
					return errors.Wrap(err, `error reading cert#S256`)
				}
				c.certS256 = &tmp
			case "jwk":
				if err := dec.Decode(&(c.jwk)); err != nil {
					return errors.Wrap(err, `error reading jwk`)
				}
			case "proof":
				if err := dec.Decode(&(c.proof)); err != nil {
					return errors.Wrap(err, `error reading proof`)
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

func (c *Key) makePairs() []*mapiter.Pair {
	var pairs []*mapiter.Pair
	if tmp := c.cert; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "cert", Value: *tmp})
	}
	if tmp := c.certS256; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "cert#S256", Value: *tmp})
	}
	if tmp := c.proof; tmp != nil {
		pairs = append(pairs, &mapiter.Pair{Key: "proof", Value: *tmp})
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

func (c *Key) Iterate(ctx context.Context) mapiter.Iterator {
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
