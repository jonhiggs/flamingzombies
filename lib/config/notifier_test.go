package config

import (
	"fmt"
	"testing"

	"github.com/jonhiggs/flamingzombies/lib/trace"
)

func TestNotifier(t *testing.T) {
	n := Notifier{}

	t.Run("initial values", func(t *testing.T) {
		if fmt.Sprintf("%v", n.envs) != fmt.Sprintf("%v", []string{}) {
			t.Errorf("got %v, want %v", n.envs, []string{})
		}
	})

	t.Run("set description", func(t *testing.T) {
		n.envs = []string{}
		n.SetDescription("the description")
		got := n.envs
		want := []string{
			"DESCRIPTION=the description",
		}
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("set message", func(t *testing.T) {
		n.envs = []string{}
		n.SetMessage("the message")
		got := n.envs
		want := []string{
			"MESSAGE=the message",
		}
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("set priority", func(t *testing.T) {
		n.envs = []string{}
		n.SetPriority(1)
		got := n.envs
		want := []string{
			"PRIORITY=1",
		}
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("set subject", func(t *testing.T) {
		n.envs = []string{}
		n.SetSubject("hello world")
		got := n.envs
		want := []string{
			"SUBJECT=hello world",
		}
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("set trace id", func(t *testing.T) {
		n.envs = []string{}
		id := trace.New()
		n.SetTraceID(id)
		got := n.envs
		want := []string{
			fmt.Sprintf("TRACE_ID=%s", id),
		}
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

}
