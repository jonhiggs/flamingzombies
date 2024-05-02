package fz

import (
	"testing"
)

func init() {
	StartLogger("info")

	config = Config{
		Notifiers: []Notifier{

			Notifier{
				Name: "zero",
				GateSets: [][]string{
					[]string{"gate_zero"},
					//[]string{}, TODO: later
				},
			},
		},
		Gates: []Gate{
			Gate{Name: "gate_zero"},
		},
	}

}

func TestNotifierByName(t *testing.T) {
	want := "zero"
	got := NotifierByName("zero")

	if got == nil {
		t.Errorf("unexpected nil notifier")
	}

	if got.Name != want {
		t.Errorf("got %s, want %s", got.Name, want)
	}

	want = "non-existent"
	got = NotifierByName("non-existent")
	if got != nil {
		t.Errorf("expected nil notifier")
	}
}

func TestNotifierGates(t *testing.T) {
	got := config.Notifiers[0].Gates()

	if len(got) != 1 {
		t.Errorf("length: got %d, want 1", len(got))
	}

	if len(got[0]) != 1 {
		t.Errorf("length 0: got %d, want 1", len(got[0]))
	}
}
