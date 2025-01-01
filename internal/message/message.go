package message

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"

	"github.com/kmpm/ged-shovel/public/models"
)

func NewZlibReader(r *bytes.Reader) (io.ReadCloser, error) {
	return zlib.NewReader(r)
}

// Deflate decompresses a zlib compressed byte slice
func Deflate(raw []byte) ([]byte, error) {

	r, err := zlib.NewReader(bytes.NewBuffer(raw))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	p := make([]byte, 1024)
	out := new(bytes.Buffer)
	n, err := r.Read(p)
	for n > 0 {
		out.Write(p[:n])
		if err != nil {
			break
		}
		n, err = r.Read(p)
	}
	if err != io.EOF {
		return nil, err
	}
	return out.Bytes(), nil

}

// DecodeReader reads a zlib compressed JSON message from an io.Reader
func DecodeReader(r io.Reader) (*models.EDDN, error) {
	z, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	var message models.EDDN
	err = json.NewDecoder(z).Decode(&message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

// Decode decodes a zlib compressed JSON message from a byte slice
func Decode(rawMessage []byte) (*models.EDDN, error) {
	return DecodeReader(bytes.NewReader(rawMessage))
}

// Encode encodes a message to a zlib compressed byte slice
func Encode(msg *models.EDDN) ([]byte, error) {
	// marshal and compress message using zlib

	zlibbed := new(bytes.Buffer)
	w := zlib.NewWriter(zlibbed)
	err := json.NewEncoder(w).Encode(msg)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return zlibbed.Bytes(), nil
}
