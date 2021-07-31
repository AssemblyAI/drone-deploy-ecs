package main

import (
	"reflect"
	"testing"
)

func Test_getServiceNames(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test-multiple-services",
			args: args{
				s: "foobar,whizbang,helloworld",
			},
			want: []string{"foobar", "whizbang", "helloworld"},
		},
		{
			name: "test-one-service",
			args: args{
				s: "helloworld",
			},
			want: []string{"helloworld"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getServiceNames(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getServiceNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
