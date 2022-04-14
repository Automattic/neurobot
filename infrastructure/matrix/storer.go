package matrix

import (
	"bytes"
	"encoding/gob"

	"github.com/apex/log"
	"github.com/upper/db/v4"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

const table = "matrix_state"

// NewStorer implements the Storer interface required by Mautrix client to store sync state persistently
func NewStorer(db db.Session, botID uint64) mautrix.Storer {
	return &storer{
		db:    db,
		botID: botID,
	}
}

// storer uses sqlite database to store state as required by Mautrix client
type storer struct {
	db    db.Session
	botID uint64
}

type row struct {
	botID uint64 `db:"bot_id,omitempty"`
	id    string // could be userID or roomID
	what  string // could be "filter", "batch", "room"
	value string
}

func (s *storer) SaveFilterID(userID id.UserID, filterID string) {
	if err := s.save("filter", userID.String(), filterID); err != nil {
		log.WithError(err).Error("SaveFilterID error")
	}
}

func (s *storer) LoadFilterID(userID id.UserID) string {
	value, err := s.get("filter", userID.String())
	if err != nil {
		log.WithError(err).Error("LoadFilterID error")
	}

	return value
}

func (s *storer) SaveNextBatch(userID id.UserID, nextBatchToken string) {
	if err := s.save("batch", userID.String(), nextBatchToken); err != nil {
		log.WithError(err).Error("SaveNextBatch error")
	}
}

func (s *storer) LoadNextBatch(userID id.UserID) string {
	value, err := s.get("batch", userID.String())
	if err != nil {
		log.WithError(err).Error("LoadNextBatch error")
	}

	return value
}

func (s *storer) SaveRoom(room *mautrix.Room) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	enc.Encode(room)

	if err := s.save("room", room.ID.String(), b.String()); err != nil {
		log.WithError(err).Error("SaveRoom error")
	}
}
func (s *storer) LoadRoom(roomID id.RoomID) *mautrix.Room {
	value, err := s.get("room", roomID.String())
	if err != nil {
		log.WithError(err).Error("LoadRoom error")
	}

	var b bytes.Buffer
	b.WriteString(value)
	dec := gob.NewDecoder(&b)
	var r mautrix.Room
	dec.Decode(&r)

	return &r
}

func (s *storer) save(what string, id string, value string) error {
	var r row
	result := s.db.Collection(table).Find(db.Cond{"bot_id": s.botID, "id": id, "type": what})
	if err := result.One(&r); err != nil {
		return err
	}

	exists, err := result.Exists()
	if err != nil {
		return err
	}

	r.value = value
	r.botID = s.botID

	if !exists {
		_, err := s.db.Collection(table).Insert(r)
		return err
	}

	return result.Update(r)
}

func (s *storer) get(what string, id string) (string, error) {
	var r row
	result := s.db.Collection(table).Find(db.Cond{"bot_id": s.botID, "id": id, "type": what})
	if err := result.One(&r); err != nil {
		return "", err
	}

	return r.value, nil
}
