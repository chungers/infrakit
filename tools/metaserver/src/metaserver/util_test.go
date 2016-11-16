package main

import "testing"

func TestConvertHostToIP(t *testing.T) {
	if convertHostToIP("ip-10-0-3-149.ec2.internal") != "10.0.3.149" {
		t.Error("Expected 10.0.3.149")
	}
	if convertHostToIP("ip-10-0-3-149") != "10.0.3.149" {
		t.Error("Expected 10.0.3.149")
	}
}
