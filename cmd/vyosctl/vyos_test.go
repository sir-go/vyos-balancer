package main

import (
	"testing"
)

func Test_foundInArrInt(t *testing.T) {
	type args struct {
		s []int
		v int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{[]int{}, 0}, false},
		{"ok-no", args{[]int{1, 2, 3, 4}, 8}, false},
		{"ok-yes", args{[]int{1, 2, 3, 5}, 2}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := foundInArrInt(tt.args.s, tt.args.v); got != tt.want {
				t.Errorf("foundInArrInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseKMGTUint(t *testing.T) {
	type args struct {
		digits     string
		multiplier string
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr bool
	}{
		{"empty", args{"", ""}, 0, true},
		{"ok-digits", args{"1234", ""}, 1234, false},
		{"ok-k", args{"12", "k"}, 12288, false},
		{"ok-m", args{"12", "m"}, 12582912, false},
		{"ok-g", args{"12", "g"}, 12884901888, false},
		{"ok-t", args{"12", "t"}, 13194139533312, false},
		{"bad-digits", args{"- d 1234", ""}, 0, true},
		{"bad-mult", args{"1234", "p"}, 1234, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKMGTUint(tt.args.digits, tt.args.multiplier)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseKMGTUint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseKMGTUint() got = %v, want %v", got, tt.want)
			}
		})
	}
}
