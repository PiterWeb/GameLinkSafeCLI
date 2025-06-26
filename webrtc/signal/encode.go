// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

// Package signal contains helpers to exchange the SDP session
// description between examples.
package signal

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
)

// Allows compressing offer/answer to bypass terminal input limits.
const compress = true

// SignalEncode encodes the input in base64
// It can optionally zip the input before encoding
func SignalEncode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	if compress {
		err, b = signalZip(b)
		if err != nil {
			panic(err)
		}
	}

	return base64.StdEncoding.EncodeToString(b)
}

// SignalDecode decodes the input from base64
// It can optionally unzip the input after decoding
func SignalDecode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	if compress {
		err, b = signalUnzip(b)
		if err != nil {
			return err
		}
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		return err
	}

	return nil
}

func signalZip(in []byte) (error, []byte) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write(in)
	if err != nil {
		return err, nil
	}
	err = gz.Flush()
	if err != nil {
		return err, nil
	}
	err = gz.Close()
	if err != nil {
		return err, nil
	}
	return nil, b.Bytes()
}

func signalUnzip(in []byte) (error, []byte) {
	var b bytes.Buffer
	_, err := b.Write(in)
	if err != nil {
		return err, nil
	}
	r, err := gzip.NewReader(&b)
	if err != nil {
		return err, nil
	}
	res, err := io.ReadAll(r)
	if err != nil {
		return err, nil
	}
	return nil, res
}
