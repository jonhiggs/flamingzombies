package fz

import (
	"testing"
)

func init() {
}

func TestConfigDefaults(t *testing.T) {
	config := ReadConfig("/home/jon/src/flamingzombies/example_config.toml")
	config.Directory = "/home/jon/src/flamingzombies/libexec"
	want := ConfigDefaults{
		Retries: 5,
	}
	got := config.Defaults

	if got.Retries != want.Retries {
		t.Errorf("got %d, want %d", got.Retries, want.Retries)
	}
}

//func TestGetNotifierByName(t *testing.T) {
//	want := "zero"
//	got := config.GetNotifierByName("zero")
//
//	if got == nil {
//		t.Errorf("unexpected nil notifier")
//	}
//
//	if got.Name != want {
//		t.Errorf("got %s, want %s", got.Name, want)
//	}
//
//	want = "non-existent"
//	got = NotifierByName("non-existent")
//	if got != nil {
//		t.Errorf("expected nil notifier")
//	}
//}
//

//func TestGetNotifierGates(t *testing.T) {
//	got := config.Notifiers[0].Gates()
//
//	if len(got) != 1 {
//		t.Errorf("length: got %d, want 1", len(got))
//	}
//
//	if len(got[0]) != 1 {
//		t.Errorf("length 0: got %d, want 1", len(got[0]))
//	}
//}
//
