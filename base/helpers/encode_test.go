package helpers

import (
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	type args struct {
		bin []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "success",
			args: args{
				bin: []byte("Hola"),
			},
			want: []byte("SG9sYQ=="),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Encode(tt.args.bin); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

//TODO: Fix this text
func TestGetExportFilename(t *testing.T) {
	type args struct {
		name     string
		mimeType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		/*{
			name: "success jpg",
			args: args{
				name:     "title",
				mimeType: "image/jpeg",
			},
			want: "title.jpeg",
		},*/
		{
			name: "success png",
			args: args{
				name:     "title",
				mimeType: "image/png",
			},
			want: "title.png",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExportFilename(tt.args.name, tt.args.mimeType); got != tt.want {
				t.Errorf("GetExportFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
