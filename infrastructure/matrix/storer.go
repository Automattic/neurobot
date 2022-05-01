package matrix

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/apex/log"
	"github.com/upper/db/v4"
	"maunium.net/go/mautrix"
	mautrixCrypto "maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

const stateTable = "matrix_state"
const membershipTable = "room_members"

type membershipTableRow struct {
	UserID string `db:"user_id"`
	RoomID string `db:"room_id"`
}

// Store is an interface that embeds both mautrix.Storer and (mautrix/crypto).StateStore
// mautrix.Storer is for general state events for bot without the context of encryption
// (mautrix/crypto).StateStore is specifically for encryption context
type Store interface {
	mautrix.Storer
	mautrixCrypto.StateStore
}

// NewStorer implements the Storer interface required by Mautrix client to store sync state persistently
func NewStorer(db db.Session, botID uint64) Store {
	return &storer{
		db:    db,
		botID: botID,
	}
}

// NewCryptoStorer implements the StateStore interface required by OlmMachine to store room state required for encryption
func NewCryptoStorer(db db.Session) mautrixCrypto.StateStore {
	return &storer{db: db}
}

// storer uses sqlite database to store state as required by Mautrix client
type storer struct {
	db    db.Session
	botID uint64
}

type stateTableRow struct {
	BotID uint64 `db:"bot_id"`
	What  string `db:"what"` // possible values are "filter", "batch", "room"
	ID    string `db:"id"`   // would be values for userID or roomID
	Value string `db:"value"`
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

// IsEncrypted returns whether a room is encrypted.
func (s *storer) IsEncrypted(roomID id.RoomID) bool {
	value, err := s.get("room_crypto_event", roomID.String())
	if err != nil || value == "" {
		return false
	}
	return true
}

// GetEncryptionEvent returns the encryption event's content for an encrypted room.
func (s *storer) GetEncryptionEvent(roomID id.RoomID) *event.EncryptionEventContent {
	value, err := s.get("room_crypto_event", roomID.String())
	if err != nil {
		// If I understand this right, it would never be queried without having it saved first - when we know room is encrypted and we are sending an encrypted message or room's encryption event was saver prior
		log.WithError(err).WithField("roomID", roomID.String()).Error("room encryption event empty")

		// should be safe to return this, but not 100% sure
		// tracking the error logged above would clear this up
		return &event.EncryptionEventContent{
			Algorithm:              id.AlgorithmMegolmV1,
			RotationPeriodMillis:   7 * 24 * 60 * 60 * 1000,
			RotationPeriodMessages: 100,
		}
	}

	var b bytes.Buffer
	b.WriteString(value)
	dec := gob.NewDecoder(&b)
	var e event.EncryptionEventContent
	dec.Decode(&e)

	return &e
}

// SetEncryptionEvent saves the encryption event's content for an encrypted room.
func (s *storer) SetEncryptionEvent(roomID id.RoomID, event *event.EncryptionEventContent) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	enc.Encode(event)

	if err := s.save("room_crypto_event", roomID.String(), b.String()); err != nil {
		log.WithError(err).WithField("roomID", roomID.String()).Error("SetEncryptionEvent error")
	}
}

// FindSharedRooms returns what other encrypted rooms the specified user has joined
func (s *storer) FindSharedRooms(userID id.UserID) []id.RoomID {
	var rows []struct {
		userID string
		roomID string
	}
	result := s.db.Collection(membershipTable).Find(db.Cond{"user_id": userID})
	err := result.All(&rows)
	if err != nil {
		if !errors.Is(err, db.ErrNoMoreRows) {
			log.WithError(err).WithField("userID", userID).Error("error fetching rooms for user")
		}
		return []id.RoomID{}
	}

	rooms := make([]id.RoomID, len(rows))
	for _, row := range rows {
		rooms = append(rooms, id.RoomID(row.roomID))
	}
	return rooms
}

// SetMembership saves the membership info that a certain user is member of a certain room
func (s *storer) SetMembership(event *event.Event) {
	roomID := event.RoomID
	userID := event.GetStateKey()

	res := s.db.Collection(membershipTable).Find(db.Cond{
		"user_id": userID,
		"room_id": roomID,
	})

	var row membershipTableRow
	row.UserID = userID
	row.RoomID = roomID.String()

	membershipEvent := event.Content.AsMember()
	if membershipEvent.Membership.IsInviteOrJoin() {
		exists, err := res.Exists()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"userID": userID,
				"roomID": roomID,
			}).Error("error querying membership data")
			return
		}
		if !exists {
			if _, err := s.db.Collection(membershipTable).Insert(row); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"userID": userID,
					"roomID": roomID,
				}).Error("error inserting membership data")
			}
		}
	} else {
		if err := res.Delete(); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"userID": userID,
				"roomID": roomID,
			}).Error("error deleting membership data")
		}
	}
}

func (s *storer) save(what string, id string, value string) error {
	var exists bool
	var r stateTableRow
	result := s.db.Collection(stateTable).Find(db.Cond{"bot_id": s.botID, "id": id, "what": what})
	if err := result.One(&r); err != nil {
		if !errors.Is(err, db.ErrNoMoreRows) {
			return err
		}
	} else {
		exists = true
	}

	// overwrite values
	r.Value = value
	r.BotID = s.botID
	r.ID = id
	r.What = what

	if !exists {
		_, err := s.db.Collection(stateTable).Insert(r)
		return err
	}

	return result.Update(r)
}

func (s *storer) get(what string, id string) (string, error) {
	var r stateTableRow
	result := s.db.Collection(stateTable).Find(db.Cond{"bot_id": s.botID, "id": id, "what": what})
	if err := result.One(&r); err != nil {
		if !errors.Is(err, db.ErrNoMoreRows) {
			return "", err
		}
	}

	return r.Value, nil
}
