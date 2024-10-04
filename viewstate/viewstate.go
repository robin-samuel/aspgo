package viewstate

import (
	"encoding/base64"
	"errors"
)

type Viewstate struct {
	raw       []byte
	decoded   any
	mac       string
	signature []byte
}

func New(data string) (*Viewstate, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return &Viewstate{raw: raw}, nil
}

func (v *Viewstate) Decode() (any, error) {
	if len(v.raw) < 2 {
		return nil, errors.New("invalid viewstate, too short")
	}
	preamble := v.raw[:2]
	if string(preamble) != "\xff\x01" {
		return nil, errors.New("invalid viewstate, bad preamble")
	}
	body := v.raw[2:]
	decoded, remain, err := parse(body)
	if err != nil {
		return nil, err
	}
	switch len(remain) {
	case 20:
		v.mac = "hmac_sha1"
		v.signature = remain
	case 32:
		v.mac = "hmac_sha256"
		v.signature = remain
	default:
		v.mac = "unknown"
		v.signature = remain
	}
	return decoded, nil
}
