package main

import "testing"

func TestCleanPath(t *testing.T) {
	if CleanPath("../../etc/passwd") != "etc/passwd" {
		t.Fatal("Did not clean path")
	}
}
