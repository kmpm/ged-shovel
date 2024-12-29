package message

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/kmpm/enp/public/models"
)

func newTestValidator(t *testing.T) *Validator {
	t.Helper()
	v, err := NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestValidator_Validate(t *testing.T) {
	type args struct {
		schemaURL string
		r         io.Reader
	}
	tests := []struct {
		name    string
		v       *Validator
		args    args
		wantErr bool
	}{
		{"Test file a", newTestValidator(t), args{"https://eddn.edcd.io/schemas/journal/1", bytes.NewReader(readFile(t, "testdata/a_deflated.dat"))}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(tt.args.schemaURL, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("Validator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func readMessage(t *testing.T, filename string) *models.EDDN {
	t.Helper()
	msg := &models.EDDN{}
	json.Unmarshal(readFile(t, filename), msg)
	return msg
}
