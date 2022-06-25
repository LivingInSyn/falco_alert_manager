package main

import (
	"testing"
)

func TestGetConfig(t *testing.T) {
	config := getConfig("test_files/config.yml")
	goodAddress := ":8081"
	if config.Server.Address != goodAddress {
		t.Fatalf("Addresses don't match. Wanted %s got %s", goodAddress, config.Server.Address)
	}
}
