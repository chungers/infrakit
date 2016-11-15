package main

import "testing"

func TestConvertAWSHostToIP(t *testing.T) {
	if ConvertAWSHostToIP("ip-10-0-3-149.ec2.internal") != "10.0.3.149" {
		t.Error("Expected 10.0.3.149")
	}
	if ConvertAWSHostToIP("ip-10-0-3-149") != "10.0.3.149" {
		t.Error("Expected 10.0.3.149")
	}
}
