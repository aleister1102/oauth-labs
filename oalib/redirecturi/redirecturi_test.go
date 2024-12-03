package redirecturi

import (
	"testing"
)

func TestRedirectURIConstruction(t *testing.T) {
	r, _ := New("http://127.0.0.1:3000/?error=123&state=123")
	r.SetError("error")
	if r.uri.Query().Get("error") != "error" {
		t.FailNow()
	}
	r.SetState("state")
	if r.uri.Query().Get("state") != "state" {
		t.FailNow()
	}
}

func TestRedirectURIConstructionErr(t *testing.T) {
	r, _ := NewWithError("http://127.0.0.1:3000/?state=123", "error")
	param := r.GetQuery("error")
	if param != "error" {
		t.FailNow()
	}
}

func TestRedirectURIConstructionState(t *testing.T) {
	r, _ := NewWithState("http://127.0.0.1:3000/?error=123", "state")
	param := r.GetQuery("state")
	if param != "state" {
		t.FailNow()
	}
}
