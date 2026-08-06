package mcpserver

import (
	"reflect"
	"testing"
)

func TestParseInfoFields(t *testing.T) {
	raw := "# Server\r\nredis_version:redis-clone-0.1\r\nrole:master\r\n# Keyspace\r\ndb0:keys=3\r\n"

	got := parseInfoFields(raw)
	want := map[string]string{
		"redis_version": "redis-clone-0.1",
		"role":          "master",
		"db0":           "keys=3",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
