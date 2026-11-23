package config

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		refresh string
		wantErr bool
	}{
		{"empty secret fails", "", "r", true},
		{"empty refresh secret fails", "s", "", true},
		{"both filled passes", "s", "r", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{JWT: JWTConfig{Secret: tt.secret, RefreshSecret: tt.refresh}}
			if err := cfg.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSplitCSV(t *testing.T) {
	got := splitCSV(" http://a.com , ,http://b.com")
	if len(got) != 2 || got[0] != "http://a.com" || got[1] != "http://b.com" {
		t.Fatalf("unexpected result: %v", got)
	}
}
