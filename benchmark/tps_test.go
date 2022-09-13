package benchmark

import "testing"

func Test_benchmarkForTps(t *testing.T) {
	type args struct {
		urls []string
	}
	tests := []struct {
		name string
		args args
	}{
		{"benchmarkForTps", args{
			urls: []string{"", ""},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			benchmarkForTps(tt.args.urls)
		})
	}
}
