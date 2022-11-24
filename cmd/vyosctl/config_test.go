package main

import (
	"testing"
	"time"
)

func TestDuration_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		dWant   time.Duration
		text    []byte
		wantErr bool
	}{
		{"empty", 0, []byte(""), true},
		{"nil", 0, nil, true},
		{"1s", time.Second, []byte("1s"), false},
		{"10m", time.Minute * 10, []byte("10m"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Duration{}
			if err := d.UnmarshalText(tt.text); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
			if d.Duration != tt.dWant {
				t.Errorf("UnmarshalText() = %v, want %v", tt.text, d.Duration.String())
			}
		})
	}
}

func TestThreshold_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		th      *Threshold
		confVal string
		wantErr bool
	}{
		{"empty", &Threshold{}, "", true},
		{"ok", &Threshold{Limit: 123}, "123", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.th.UnmarshalText([]byte(tt.confVal)); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
