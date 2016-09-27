package ratelimit

import (
	"github.com/wscherphof/msg"
)

func Init() {
	msg.New("ErrTooManyRequests").
		Add("nl", `Sorry; hier geldt een frequentielimiet; je kan dit maar eens in
      de zoveel tijd aanvragen.
      Probeer het later opnieuw.`).
		Add("en", `Sorry; a rate limit is in effect for this request type.
      Please try again later.`)

	msg.New("ErrInvalidRequest").
		Add("nl", `Sorry; er klopt iets niet in het kader van de frequentielimiet
      die voor dit verzoek van kracht is.`).
		Add("en", `Sorry; something is wrong with the rate limit controls.`)

	msg.New("Limit is set to").
		Add("nl", `De limiet is ingesteld op:`).
		Add("en", `Rate limit is set to:`)
}
