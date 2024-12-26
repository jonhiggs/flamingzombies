package fz

import (
	"fmt"
	"testing"
)

func init() {
}

func TestConfigDefaults(t *testing.T) {
	config := ReadConfig("/home/jon/src/flamingzombies/example_config.toml")
	config.Directory = "/home/jon/src/flamingzombies/libexec"
	want := ConfigDefaults{
		Retries:            5,
		TimeoutSeconds:     1,
		NotifierNames:      []string{"logger"},
		ErrorNotifierNames: []string{"error_emailer"},
		Priority:           3,
		FrequencySeconds:   0,
		TaskEnvs: [][]string{
			[]string{"SNMP_COMMUNITY", "default"},
			[]string{"SNMP_VERSION", "2c"},
		},
	}
	got := config.Defaults

	if got.Retries != want.Retries {
		t.Errorf("got %d, want %d", got.Retries, want.Retries)
	}

	if got.TimeoutSeconds != want.TimeoutSeconds {
		t.Errorf("got %d, want %d", got.TimeoutSeconds, want.TimeoutSeconds)
	}

	if fmt.Sprintf("%s", got.NotifierNames) != fmt.Sprintf("%s", want.NotifierNames) {
		t.Errorf("got %v, want %v", got.NotifierNames, want.NotifierNames)
	}

	if fmt.Sprintf("%s", got.ErrorNotifierNames) != fmt.Sprintf("%s", want.ErrorNotifierNames) {
		t.Errorf("got %v, want %v", got.ErrorNotifierNames, want.ErrorNotifierNames)
	}

	if got.Priority != want.Priority {
		t.Errorf("got %d, want %d", got.Priority, want.Priority)
	}

	if fmt.Sprintf("%v", got.TaskEnvs) != fmt.Sprintf("%s", want.TaskEnvs) {
		t.Errorf("got %v, want %v", got.TaskEnvs, want.TaskEnvs)
	}

	if got.FrequencySeconds != want.FrequencySeconds {
		t.Errorf("got %d, want %d", got.FrequencySeconds, want.FrequencySeconds)
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
