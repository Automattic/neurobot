package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/upper/db/v4"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Bot struct {
	ID          uint64 `db:"id,omitempty"`
	Identifier  string `db:"identifier"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Username    string `db:"username"`
	Password    string `db:"password"`
	CreatedBy   string `db:"created_by"`
	Active      bool   `db:"active"`

	hydrated bool

	e *engine
}

func getActiveBots(dbs db.Session) (bots []Bot, err error) {
	res := dbs.Collection("bots").Find(db.Cond{"active": 1})
	err = res.All(&bots)
	if err != nil {
		return
	}

	return
}

func getBot(dbs db.Session, identifier string) (b Bot, err error) {
	res := dbs.Collection("bots").Find(db.Cond{"identifier": identifier})
	err = res.One(&b)

	return
}

func (b *Bot) IsHydrated() bool {
	return b.hydrated
}

func (b *Bot) Hydrate(e *engine) {
	b.hydrated = true
	b.e = e
}

func (b *Bot) WakeUp(e *engine) (err error) {
	// save reference
	b.e = e

	b.log(fmt.Sprintf("Matrix: Activating Bot: %s [%s]", b.Name, b.Identifier))

	client, err := mautrix.NewClient(e.matrixhomeserver, "", "")
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
	b.log(fmt.Sprintf("Matrix: Bot %s [%s] login successful", b.Name, b.Identifier))

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
			matrixHSHost := strings.Split(strings.Split(b.e.matrixhomeserver, "://")[1], ":")[0] // remove protocol and port info to get just the hostname
			if strings.Split(evt.RoomID.String(), ":")[1] == matrixHSHost {
				// join the room
				_, err := b.JoinRoom(evt.RoomID)
				if err != nil {
					b.log(fmt.Sprintf("Bot couldn't join the invitation bot:%s invitation:%s", b.Name, evt.RoomID))
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

func (b *Bot) JoinRoom(roomid id.RoomID) (resp *mautrix.RespJoinRoom, err error) {
	if c := b.getMCInstance(); c != nil {
		return c.JoinRoomByID(roomid)
	}

	return nil, errors.New("bot instance not hydrated")
}

func (b *Bot) log(m string) {
	b.e.log(m)
}
