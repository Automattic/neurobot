package engine

import (
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

func (b Bot) WakeUp(e *engine) (client MatrixClient, err error) {
	e.log(fmt.Sprintf("Matrix: Activating Bot: %s [%s]", b.Name, b.Identifier))

	// save reference
	b.e = e

	client, err = mautrix.NewClient(e.matrixhomeserver, "", "")
	if err != nil {
		return
	}
	_, err = client.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: b.Username},
		Password:         b.Password,
		StoreCredentials: true,
	})
	if err != nil {
		return
	}
	e.log(fmt.Sprintf("Matrix: Bot %s [%s] login successful", b.Name, b.Identifier))

	syncer := client.(*mautrix.Client).Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.StateMember, b.HandleStateMemberEvent)

	err = client.Sync()

	return
}

func (b Bot) HandleStateMemberEvent(source mautrix.EventSource, evt *event.Event) {
	if membership, ok := evt.Content.Raw["membership"]; ok {
		if membership == "invite" {
			b.log(fmt.Sprintf("Invitation for %s\n", evt.RoomID))

			// ensure the invitation if for a room within our homeserver only
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

func (b Bot) getInstance() MatrixClient {
	return b.e.bots[b.ID]
}

func (b Bot) JoinRoom(roomid id.RoomID) (resp *mautrix.RespJoinRoom, err error) {
	return b.getInstance().JoinRoomByID(roomid)
}

func (b Bot) log(m string) {
	b.e.log(m)
}
