package message

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestDeflate(t *testing.T) {
	type args struct {
		raw []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"Test file a", args{readFile(t, "testdata/a_compressed.dat")}, readFile(t, "testdata/a_deflated.dat"), false},
		{"Test file b", args{readFile(t, "testdata/b_compressed.dat")}, readFile(t, "testdata/b_deflated.dat"), false},
		{"Test file c", args{readFile(t, "testdata/c_compressed.dat")}, readFile(t, "testdata/c_deflated.dat"), false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Deflate(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deflate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !assert.Equal(t, tt.want, got) {
				os.WriteFile("testdata/tmp_"+tt.name+"_got.dat", got, 0644)
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Deflate() = %v, want %v", string(got), string(tt.want))
			// }
		})
	}
}

func TestDecodeReader(t *testing.T) {

	tests := []struct {
		name      string
		inputfile string
		wantfile  string
		wantErr   bool
	}{
		{"Test file a", "testdata/a_compressed.dat", "testdata/a.json", false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inb, err := os.Open(tt.inputfile)
			if err != nil {
				t.Fatal(err)
			}
			defer inb.Close()
			got, err := DecodeReader(inb)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotData, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			wantData := readFile(t, tt.wantfile)
			if !reflect.DeepEqual(gotData, wantData) {
				t.Errorf("DecodeReader() = %v, want %v", string(gotData), string(wantData))
				// os.WriteFile("testdata/"+tt.name+".json", gotData, 0644)
			}
		})
	}
}
