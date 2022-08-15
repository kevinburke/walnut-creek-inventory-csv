package main

import "testing"

func TestZoningRx(t *testing.T) {
	if !zoningRx.MatchString("P-D") {
		t.Errorf("should have matched P-D, did not")
	}
	if !zoningRx.MatchString("MU-0.75") {
		t.Errorf("should have matched P-D, did not")
	}
}
