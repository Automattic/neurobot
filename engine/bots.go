package engine

import (
	"fmt"
	"strings"

	"github.com/upper/db/v4"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
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

func (b Bot) WakeUp(e *engine) error {
	e.log(fmt.Sprintf("Matrix: Activating Bot: %s [%s]", b.Name, b.Identifier))

	client, err := mautrix.NewClient(e.matrixhomeserver, "", "")
	if err != nil {
		return err
	}
	_, err = client.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: b.Username},
		Password:         b.Password,
		StoreCredentials: true,
	})
	if err != nil {
		return err
	}
	e.log(fmt.Sprintf("Matrix: Bot %s [%s] login successful", b.Name, b.Identifier))

	syncer := client.Syncer.(*mautrix.DefaultSyncer)

	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		if membership, ok := evt.Content.Raw["membership"]; ok {
			if membership == "invite" {
				e.log(fmt.Sprintf("Invitation for %s\n", evt.RoomID))

				// ensure the invitation if for a room within our homeserver only
				matrixHSHost := strings.Split(strings.Split(e.matrixhomeserver, "://")[1], ":")[0] // remove protocol and port info to get just the hostname
				if strings.Split(evt.RoomID.String(), ":")[1] == matrixHSHost {
					// join the room
					_, err := client.JoinRoomByID(evt.RoomID)
					if err != nil {
						e.log(fmt.Sprintf("Bot couldn't join the invitation bot:%s invitation:%s", b.Name, evt.RoomID))
					} else {
						e.log("accepted invitation, if it wasn't accepted already")
					}
				} else {
					e.log(fmt.Sprintf("whaat? %v", strings.Split(evt.RoomID.String(), ":")))
				}
			}
		}
		e.log(fmt.Sprintf("\nSource: %d\n%s  %s\n%+v\n", source, evt.Type.Type, evt.RoomID, evt.Content.Raw))
	})

	return client.Sync()
}
