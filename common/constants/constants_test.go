package constants

import "testing"

func TestDriverGeoKeyOf(t *testing.T) {
	if got := DriverGeoKeyOf(""); got != DriverGeoKey {
		t.Fatalf("empty city should fallback to default key, got %s", got)
	}
	if got := DriverGeoKeyOf("110100"); got != "driver:geo:110100" {
		t.Fatalf("city key mismatch, got %s", got)
	}
}

func TestOrderStatusValues(t *testing.T) {
	if OrderStatusWaitAccept != 1 || OrderStatusCompleted != 5 {
		t.Fatal("order status constants mismatch")
	}
}
