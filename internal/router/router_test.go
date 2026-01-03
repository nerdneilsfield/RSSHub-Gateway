package router

import "testing"

func TestSelectLongestPrefix(t *testing.T) {
	r := New([]Route{
		{Name: "a", Allow: []string{"/telegram/"}, Priority: 1, Order: 0},
		{Name: "b", Allow: []string{"/telegram/private/"}, Priority: 1, Order: 1},
	}, "default")

	got := r.Select("/telegram/private/x")
	if got != "b" {
		t.Fatalf("expected b, got %s", got)
	}
}

func TestSelectDenyOverrides(t *testing.T) {
	r := New([]Route{
		{Name: "a", Allow: []string{"/"}, Deny: []string{"/private/"}, Priority: 1, Order: 0},
		{Name: "b", Allow: []string{"/"}, Priority: 1, Order: 1},
	}, "default")

	got := r.Select("/private/x")
	if got != "b" {
		t.Fatalf("expected b, got %s", got)
	}
}

func TestSelectPriority(t *testing.T) {
	r := New([]Route{
		{Name: "a", Allow: []string{"/"}, Priority: 1, Order: 0},
		{Name: "b", Allow: []string{"/"}, Priority: 10, Order: 1},
	}, "default")

	got := r.Select("/any")
	if got != "b" {
		t.Fatalf("expected b, got %s", got)
	}
}

func TestSelectDefault(t *testing.T) {
	r := New([]Route{
		{Name: "a", Allow: []string{"/x/"}, Priority: 1, Order: 0},
	}, "default")

	got := r.Select("/other")
	if got != "default" {
		t.Fatalf("expected default, got %s", got)
	}
}
