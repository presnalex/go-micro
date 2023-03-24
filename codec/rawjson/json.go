// Package rawjson provides a rawjson codec
package rawjson

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"unicode/utf8"

	oldproto "github.com/golang/protobuf/proto"

	raw "github.com/presnalex/codec-bytes"
	"github.com/segmentio/encoding/json"
	"go.unistack.org/micro/v3/broker"
	"go.unistack.org/micro/v3/codec"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Codec struct{}

func (c *Codec) ReadHeader(reader io.Reader, m *codec.Message, t codec.MessageType) error {
	return nil
}

func (c *Codec) ReadBody(reader io.Reader, b interface{}) error {
	if b == nil {
		return nil
	}

	if frame, ok := b.(*raw.Frame); ok {
		buf, err := ioutil.ReadAll(reader)
		frame.Data = buf
		return err
	}

	if pb, ok := b.(proto.Message); ok {
		buf, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		return jsonpb.Unmarshal(buf, pb)
	}

	dec := json.NewDecoder(reader)
	dec.ZeroCopy()

	return dec.Decode(b)
}

func (c *Codec) Write(writer io.Writer, m *codec.Message, b interface{}) error {
	if b == nil {
		return nil
	}

	if frame, ok := b.(*raw.Frame); ok {
		_, err := writer.Write(frame.Data)
		return err
	}
	enc := json.NewEncoder(writer)
	enc.SetEscapeHTML(false)
	enc.SetSortMapKeys(false)
	return enc.Encode(b)
}

func NewCodec() codec.Codec {
	c := &Codec{}

	return c
}

func (c *Codec) Marshal(v interface{}, opts ...codec.Option) ([]byte, error) {
	var err error
	var ret []byte

	switch m := v.(type) {
	case *raw.Frame:
		return m.Data, nil
	case proto.Message:
		ret, err = JSONPbMarshaler.Marshal(m)
		return ret, err
	case oldproto.Message:
		b := bytes.NewBuffer(nil)
		err = OldJSONPbMarshaler.Marshal(b, m)
		ret = b.Bytes()
		return ret, err
	case *broker.Message:
		break
	default:
		return json.Marshal(v)
	}

	bm, ok := v.(*broker.Message)
	if !ok {
		return nil, fmt.Errorf("invalid message: %v", v)
	}

	// this is not go-micro case, so pass as-is
	if len(bm.Header) == 0 {
		return bm.Body, nil
	}

	switch bm.Header["Content-Type"] {
	// pass bytes as-is
	case "application/bytes-plain":
		return bm.Body, nil
		// guard from protobuf encoded message, skip processing
	case "application/grpc+proto":
		if ret, err = json.Append(ret[:0], v, 0); err != nil {
			return nil, err
		}
		return ret, nil
	}

	dst := string(bm.Body)
	nm := &Message{}
	if str, err := strconv.Unquote(string(bm.Body)); err == nil {
		dst = str
	}

	if b64, err := base64.StdEncoding.DecodeString(dst); err == nil && utf8.Valid(b64) {
		dst = string(b64)
	}

	nm.Body = []byte(strconv.Quote(dst))
	nm.Header = make(map[string]string, len(bm.Header))
	for k, v := range bm.Header {
		if !strings.Contains(v, "Bearer") {
			nm.Header[k] = v
		}
	}

	ret, err = json.Append(ret[:0], nm, 0)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *Codec) Unmarshal(d []byte, v interface{}, opts ...codec.Option) error {
	if frame, ok := v.(*raw.Frame); ok {
		frame.Data = d
		return nil
	}

	switch m := v.(type) {
	case proto.Message:
		return JSONPbUnmarshaler.Unmarshal(d, m)
	case oldproto.Message:
		return OldJSONPbUnmarshaler.Unmarshal(bytes.NewReader(d), m)
	}

	bm, ok := v.(*broker.Message)
	if !ok {
		return fmt.Errorf("invalid message: %v", v)
	}

	nm := &Message{}
	if _, err := json.Parse(d, nm, json.ZeroCopy); err != nil {
		return err
	}

	// guard from protobuf encoded message or raw.Frame, skip processing
	if nm.Header["Content-Type"] == "application/grpc+proto" {
		if _, err := json.Parse(d, v, json.ZeroCopy); err != nil {
			return err
		}
		return nil
	}

	dst := string(nm.Body)

	if str, err := strconv.Unquote(dst); err == nil {
		dst = str
	}

	if b64, err := base64.StdEncoding.DecodeString(dst); err == nil && utf8.Valid(b64) {
		dst = string(b64)
	}

	bm.Body = []byte(dst)
	bm.Header = make(map[string]string, len(nm.Header))
	for k, v := range nm.Header {
		if !strings.Contains(v, "Bearer") {
			bm.Header[k] = v
		}
	}

	return nil
}

func (c *Codec) String() string {
	return "rawjson"
}

func (c *Codec) Name() string {
	return "rawjson"
}
