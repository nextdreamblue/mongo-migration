package main

import (
	"testing"
)

func TestValidSession(t *testing.T) {
	Session, err := getSession("mongodb://localhost:27017/example")
	if err != nil {
		t.Error("Expected to connect on localhost, but: ", err)
	}
	if Session == nil {
		t.Error("Expected session to exist, got nil :(")
	}
}
func TestInValidSession(t *testing.T) {
	_, err := getSession("invalid-URL")
	if err == nil {
		t.Error("Expected to not connect on crazy env")
	}
}