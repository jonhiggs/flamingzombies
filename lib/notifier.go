package fz

import "fmt"

type Notifier struct {
	Name    string
	Command string
	Args    []string
}

type Notification struct {
	Notifier Notifier
	Subject  string
	Body     string
}

var NotifyCh = make(chan Notification, 100)

func ProcessNotifications() {
	go func() {
		for {
			select {
			case m := <-NotifyCh:
				fmt.Println("notifying", m.Notifier.Name)
			}
		}
	}()
}
