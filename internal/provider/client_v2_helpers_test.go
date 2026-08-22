package provider

import "testing"

func TestAdvanceListPage(t *testing.T) {
	two := int32(2)
	t.Run("follows next-page when present", func(t *testing.T) {
		got, more := advanceListPage(1, 100, 100, &two)
		if !more || got != 2 {
			t.Fatalf("got (%d, %v), want (2, true)", got, more)
		}
	})
	t.Run("walks a full page when next-page is missing", func(t *testing.T) {
		got, more := advanceListPage(1, 100, 100, nil)
		if !more || got != 2 {
			t.Fatalf("got (%d, %v), want (2, true)", got, more)
		}
	})
	t.Run("stops on a short page when next-page is missing", func(t *testing.T) {
		_, more := advanceListPage(2, 100, 5, nil)
		if more {
			t.Fatal("expected no further page")
		}
	})
}
