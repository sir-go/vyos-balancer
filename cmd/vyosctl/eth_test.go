package main

import (
	"reflect"
	"testing"
)

func TestThreshold_CheckLimit(t1 *testing.T) {
	tests := []struct {
		name              string
		overflowCount     uint
		th                *Threshold
		val               int
		wantOverflowCount uint
		wantOverflowed    bool
	}{
		{"empty", 0, &Threshold{Limit: 0}, 0, 0, false},
		{"ok-10-5", 3, &Threshold{Limit: 10}, 5, 0, false},
		{"ok-10-15", 3, &Threshold{Limit: 10}, 15, 1, false},
		{"ok-5-5", 1, &Threshold{Limit: 5}, 5, 0, false},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			tt.th.CheckLimit(tt.val, tt.overflowCount)
			if tt.th.Overflowed != tt.wantOverflowed {
				t1.Errorf("after t.CheckLimit(), t.Overflowed = %v, want %v",
					tt.th.Overflowed, tt.wantOverflowed)
			}
			if tt.th.OverflowCount != tt.wantOverflowCount {
				t1.Errorf("after t.CheckLimit(), t.OverflowCount = %v, want %v",
					tt.th.OverflowCount, tt.wantOverflowCount)
			}
		})
	}
}

func TestReadLines(t *testing.T) {
	want := []string{
		"Inter-|   Receive                                                |  Transmit",
		" face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed",
		"lo: 25808277   79107    0    0    0     0          0         0 25808277   79107    0    0    0     0       0          0",
		"enp1s0:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0",
		"virbr0:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0",
		"enp3s0.304: 3432411134 34553    0    0    0     0          0         0 743454  3456    0    0    0     0       0          0",
		"eno1.100: 123112311 34572    0    0    0     0          0         0 32456  3535    0    0    0     0       0          0",
		"wlp2s0: 1811762113 1525692    0    0    0     0          0         0 137621020  728210    0    0    0     0       0          0",
		"docker0:  500872    7765    0    0    0     0          0         0 26407525   16289    0    0    0     0       0          0",
		"br-e1366832ce7e:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0",
	}
	t.Run("i9n", func(t *testing.T) {
		if got := ReadLines("../../testdata/proc-dev-net"); !reflect.DeepEqual(got, want) {
			t.Errorf("ReadLines() = %v, want %v", got, want)
		}
	})
}

func TestGetByAlias(t *testing.T) {
	found := &Uplink{Alias: "eth1"}
	type args struct {
		ups   []*Uplink
		alias string
	}
	tests := []struct {
		name string
		args args
		want *Uplink
	}{
		{"empy", args{[]*Uplink{}, ""}, nil},
		{"no", args{[]*Uplink{{Alias: "eth0"}, {Alias: "eth1"}}, "eno1"}, nil},
		{"yes", args{[]*Uplink{{Alias: "eth0"}, found}, "eth1"}, found},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetByAlias(tt.args.alias, tt.args.ups...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetByAlias() = %v, want %v", got, tt.want)
			}
		})
	}
}
