package net

import (
	"reflect"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/log"
)

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    *Address
		wantErr bool
	}{
		{
			address: "localhost",
			want:    &Address{Host: "localhost"},
			wantErr: false,
		},
		{
			address: "localhost:0",
			want:    nil,
			wantErr: true,
		},
		{
			address: "localhost:-1",
			want:    nil,
			wantErr: true,
		},
		{
			address: ":12",
			want:    &Address{Host: "", Port: 12, HasPort: true},
			wantErr: false,
		},
		{
			address: "www.google.com:80",
			want:    &Address{Host: "www.google.com", Port: 80, HasPort: true},
			wantErr: false,
		},
		{
			address: "[::1]:80",
			want:    &Address{Host: "::1", Port: 80, HasPort: true, IsIPv6: true},
			wantErr: false,
		},
		{
			address: "2001:cdba:0000:0000:0000:0000:3257:9652",
			want:    &Address{Host: "2001:cdba:0000:0000:0000:0000:3257:9652", IsIPv6: true},
			wantErr: false,
		},
		{
			address: "[2001:cdba:0000:0000:0000:0000:3257:9652]",
			want:    &Address{Host: "2001:cdba:0000:0000:0000:0000:3257:9652", IsIPv6: true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAddress(tt.address, log.NewStackLogger())
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAddress('%s') error = %v, wantErr %v", tt.address, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAddress('%s') = %+v, want %+v", tt.address, got, tt.want)
			}
		})
	}
}
