package messages

import (
	"github.com/wscherphof/msg"
)

func Load() {
	msg.New("").
		Add("nl", "").
		Add("en", "")

	msg.New("Hi").
		Add("nl", "Hoi").
		Add("en", "Hi")

	msg.New("Time format").
		Add("nl", "2-1-2006, 15.04 u. (MST)").
		Add("en", "1/2/2006 at 3:04pm (MST)")

	msg.New("Last modified").
		Add("nl", "Laatste wijziging").
		Add("en", "Last modified")

	msg.New("Save changes").
		Add("nl", "Wijzigingen opslaan").
		Add("en", "Save changes")

	msg.New("Current").
		Add("nl", "Huidig").
		Add("en", "Current")

	msg.New("New").
		Add("nl", "Nieuw").
		Add("en", "New")

	msg.New("change").
		Add("nl", "wijzigen").
		Add("en", "change")

	msg.New("Confirm").
		Add("nl", "Bevestigen").
		Add("en", "Confirm")
}