package engine

import (
	"errors"
	"fmt"
	model "neurobot/model/bot"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

type Bot struct {
	ID          uint64 `db:"id,omitempty"`
	Description string `db:"description"`
	Username    string `db:"username"`
	Password    string `db:"password"`
	Active      bool   `db:"active"`

	hydrated bool

	e *engine
}

func MakeBotFromModelBot(bot model.Bot) Bot {
	return Bot{
		ID:          bot.ID,
		Description: bot.Description,
		Username:    bot.Username,
		Password:    bot.Password,
		Active:      bot.Active,
	}
}

func (b *Bot) IsHydrated() bool {
	return b.hydrated
}

func (b *Bot) Hydrate(e *engine) {
	b.hydrated = true
	b.e = e
}

func (b *Bot) WakeUp(e *engine) (err error) {
	// hydrate bot
	b.Hydrate(e)

	b.log(fmt.Sprintf("Matrix: Activating Bot: %s", b.Username))
	client, err := mautrix.NewClient(e.matrixServerURL, "", "")
	if err != nil {
		return
	}

	// save reference
	e.bots[b.ID] = client

	_, err = client.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: b.Username},
		Password:         b.Password,
		DeviceID:         "NEUROBOT",
		StoreCredentials: true,
	})
	if err != nil {
		return
	}
	b.log(fmt.Sprintf("Matrix: Bot %s login successful", b.Username))

	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.StateMember, b.HandleStateMemberEvent)

	// Fire 'sync' in another go routine since its blocking
	go func() {
		e.log(client.Sync().Error())
	}()

	return
}

func (b *Bot) HandleStateMemberEvent(source mautrix.EventSource, evt *event.Event) {
	if membership, ok := evt.Content.Raw["membership"]; ok {
		if membership == "invite" {
			b.log(fmt.Sprintf("Invitation for %s\n", evt.RoomID))

			// ensure the invitation is for a room within our homeserver only
			matrixHSHost := strings.Split(b.e.matrixServerName, ":")[0] // remove protocol and port info to get just the hostname
			if strings.Split(evt.RoomID.String(), ":")[1] == matrixHSHost {
				// join the room
				_, err := b.JoinRoom(evt.RoomID.String())
				if err != nil {
					b.log(fmt.Sprintf("Bot couldn't join the invitation bot:%s invitation:%s err:%s", b.Username, evt.RoomID, err))
				} else {
					b.log("accepted invitation, if it wasn't accepted already")
				}
			} else {
				b.log(fmt.Sprintf("whaat? %v", strings.Split(evt.RoomID.String(), ":")))
			}
		}
	}
	b.log(fmt.Sprintf("\nSource: %d\n%s  %s\n%+v\n", source, evt.Type.Type, evt.RoomID, evt.Content.Raw))
}

func (b *Bot) getMCInstance() MatrixClient {
	if b.IsHydrated() {
		return b.e.bots[b.ID]
	}

	return nil
}

func (b *Bot) JoinRoom(roomIDorAlias string) (resp *mautrix.RespJoinRoom, err error) {
	if c := b.getMCInstance(); c != nil {
		return c.JoinRoom(roomIDorAlias, "", "")
	}

	return nil, errors.New("bot instance not hydrated")
}

func (b *Bot) log(m string) {
	b.e.log(m)
}
