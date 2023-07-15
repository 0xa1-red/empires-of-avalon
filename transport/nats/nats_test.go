package nats

import (
	"fmt"
	"testing"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/spf13/viper"
)

func TestBuildURL(t *testing.T) {
	tests := []struct {
		host            string
		port            string
		user            string
		passwd          string
		expectedConnect string
		expectedString  string
	}{
		{
			expectedConnect: "nats://127.0.0.1:4222",
		},
		{
			host:            "192.168.1.1",
			port:            "6222",
			expectedConnect: "nats://192.168.1.1:6222",
		},
		{
			host:            "192.168.1.1",
			port:            "6222",
			user:            "testuser",
			expectedConnect: "nats://testuser@192.168.1.1:6222",
		},
		{
			host:            "192.168.1.1",
			port:            "6222",
			user:            "testuser",
			passwd:          "testpassword",
			expectedConnect: "nats://testuser:testpassword@192.168.1.1:6222",
			expectedString:  "nats://testuser:*****@192.168.1.1:6222",
		},
	}

	config.Setup("")

	for i, tt := range tests {
		tf := func(t *testing.T) {
			if tt.host != "" {
				viper.Set(config.NATS_Host, tt.host)
			}
			if tt.port != "" {
				viper.Set(config.NATS_Port, tt.port)
			}
			if tt.user != "" {
				viper.Set(config.NATS_User, tt.user)
			}
			if tt.passwd != "" {
				viper.Set(config.NATS_Password, tt.passwd)
			}

			u := buildURL()
			if actual, expected := u.connect(), tt.expectedConnect; actual != expected {
				t.Fatalf("FAIL: expected %s, got %s", expected, actual)
			}

			if tt.expectedString != "" {
				if actual, expected := u.String(), tt.expectedString; actual != expected {
					t.Fatalf("FAIL: expected %s, got %s", expected, actual)
				}
			}
		}

		t.Run(fmt.Sprintf("case_%d", i), tf)
	}
}
