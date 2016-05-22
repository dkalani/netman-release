package models

import (
	"encoding/hex"
	"errors"
)

type PacketTag []byte

func (pt PacketTag) String() string {
	return hex.EncodeToString(pt)
}

func (pt *PacketTag) MarshalJSON() ([]byte, error) {
	dst := make([]byte, hex.EncodedLen(len(*pt))+2)
	hex.Encode(dst[1:], *pt)
	dst[0] = byte('"')
	dst[len(dst)-1] = byte('"')
	return dst, nil
}

func (pt *PacketTag) UnmarshalJSON(jsonEncoded []byte) error {
	if jsonEncoded[0] != byte('"') {
		return errors.New("unmarshal PacketTag: missing leading double-quote")
	}
	if jsonEncoded[len(jsonEncoded)-1] != byte('"') {
		return errors.New("unmarshal PacketTag: missing trailing double-quote")
	}
	hexEncoded := jsonEncoded[1 : len(jsonEncoded)-1]
	decoded := make([]byte, hex.DecodedLen(len(hexEncoded)))
	_, err := hex.Decode(decoded, hexEncoded)
	if err != nil {
		return err
	}
	*pt = decoded
	return nil
}

func PT(s string) *PacketTag {
	b := PacketTag([]byte(s))
	return &b
}

type TaggedGroup struct {
	ID  string     `json:"id"`
	Tag *PacketTag `json:"tag"`
}

type IngressWhitelist struct {
	Destination    TaggedGroup   `json:"destination"`
	AllowedSources []TaggedGroup `json:"allowed_sources"`
}
