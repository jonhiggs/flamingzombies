package fz

import (
	"fmt"
	"testing"
	"time"
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

func TestConfigTaskFlappy(t *testing.T) {
	config := ReadConfig("/home/jon/src/flamingzombies/example_config.toml")
	config.Directory = "/home/jon/src/flamingzombies/libexec"
	want := Task{
		Name:             "flappy",
		Command:          "task/flappy",
		FrequencySeconds: 1,
		ErrorBody:        "flappy has entered an error state\n",
		RecoverBody:      "flappy has recovered\n",
	}
	got := config.Tasks[0]

	if got.Name != want.Name {
		t.Errorf("got %s, want %s", got.Name, want.Name)
	}

	if got.Command != want.Command {
		t.Errorf("got %s, want %s", got.Command, want.Command)
	}

	if got.FrequencySeconds != want.FrequencySeconds {
		t.Errorf("got %d, want %d", got.FrequencySeconds, want.FrequencySeconds)
	}

	if got.Frequency() != time.Duration(want.FrequencySeconds)*time.Second {
		t.Errorf("got %d, want %d", got.Frequency(), time.Duration(want.FrequencySeconds)*time.Second)
	}

	if got.ErrorBody != want.ErrorBody {
		t.Errorf("got %s, want %s", got.ErrorBody, want.ErrorBody)
	}

	if got.RecoverBody != want.RecoverBody {
		t.Errorf("got %s, want %s", got.RecoverBody, want.RecoverBody)
	}
}

func TestConfigNotifierLogger(t *testing.T) {
	config := ReadConfig("/home/jon/src/flamingzombies/example_config.toml")
	config.Directory = "/home/jon/src/flamingzombies/libexec"
	want := Notifier{
		Name:           "logger",
		Command:        "notifier/null",
		TimeoutSeconds: 5,
	}
	got := config.Notifiers[0]

	if got.Name != want.Name {
		t.Errorf("got %s, want %s", got.Name, want.Name)
	}

	if got.Command != want.Command {
		t.Errorf("got %s, want %s", got.Command, want.Command)
	}

	if got.TimeoutSeconds != want.TimeoutSeconds {
		t.Errorf("got %d, want %d", got.TimeoutSeconds, want.TimeoutSeconds)
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
