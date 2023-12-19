package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/brianaung/rtm/view"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Room struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"roomname"`
	CreatorID uuid.UUID `json:"creator_id"`
}

type RoomUser struct {
	RoomID uuid.UUID `json:"room_id"`
	UserID uuid.UUID `json:"user_id"`
}

type Message struct {
	ID     uuid.UUID `json:"id"`
	Msg    string    `json:"msg"`
	Time   time.Time `json:"time"`
	RoomID uuid.UUID `json:"room_id"`
	UserID uuid.UUID `json:"user_id"`
}

// =================================== Creating rooms and adding users to rooms ===================================
// createRoomWithCreator adds a new room entry and add the creator as an initial member of the room.
//
// This function first creates a new entry in the room table, then it updates
// the room_user table with the creator_id so that the creator will be apart of
// the newly created room.
func createRoomWithCreator(ctx context.Context, db *pgxpool.Pool, r *Room) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// add room entry (update room table)
	if err := addRoomEntry(ctx, tx, r); err != nil {
		return err
	}
	// add user to room (update room_user table)
	if err := addUserRoomEntry(ctx, tx, &RoomUser{RoomID: r.ID, UserID: r.CreatorID}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func addUserToRoom(ctx context.Context, db *pgxpool.Pool, ru *RoomUser) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := addUserRoomEntry(ctx, tx, &RoomUser{RoomID: ru.RoomID, UserID: ru.UserID}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func addRoomEntry(ctx context.Context, tx pgx.Tx, r *Room) error {
	_, err := tx.Exec(ctx, `insert into room(id, roomname, creator_id) values($1, $2, $3)`, r.ID, r.Name, r.CreatorID)
	return err
}

func addUserRoomEntry(ctx context.Context, tx pgx.Tx, ru *RoomUser) error {
	_, err := tx.Exec(ctx, `insert into room_user(room_id, user_id) values($1, $2)`, ru.RoomID, ru.UserID)
	return err
}

func isAMember(ctx context.Context, db *pgxpool.Pool, ru *RoomUser) (bool, error) {
	exists := false
	err := db.QueryRow(ctx, `select exists(select 1 from room_user ru where ru.room_id = $1 and ru.user_id = $2)`, ru.RoomID, ru.UserID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// ================================================================================================================

func addMessageEntry(ctx context.Context, db *pgxpool.Pool, m *Message) error {
	_, err := db.Exec(ctx, `insert into message(id, msg, time, room_id, user_id) values($1, $2, $3, $4, $5)`, m.ID, m.Msg, m.Time, m.RoomID, m.UserID)
	return err
}

func getRoomByID(ctx context.Context, db *pgxpool.Pool, rid uuid.UUID) (*Room, error) {
	r := &Room{}
	err := db.QueryRow(ctx, `select * from room where room.id = $1`, rid).Scan(&r.ID, &r.Name, &r.CreatorID)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func getAllRooms(ctx context.Context, db *pgxpool.Pool) ([]*Room, error) {
	rows, err := db.Query(ctx, `select * from room`)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	rooms := make([]*Room, 0)
	for rows.Next() {
		room := &Room{}
		err := rows.Scan(&room.ID, &room.Name, &room.CreatorID)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// getMessagesFromRoom retrieves top 10 latest messages from the specified room.
//
// The data is formatted in a way to use as view.MsgData by the relevant html templates.
func getMessagesFromRoom(ctx context.Context, db *pgxpool.Pool, rid uuid.UUID, uid uuid.UUID) ([]view.MsgDisplayData, error) {
	rows, err := db.Query(ctx,
		`select message.msg, message.time, u.username, message.user_id = $1 as mine
            from message 
            inner join "user" u on u.id = message.user_id
            where message.room_id = $2
            order by message.time desc
            limit 10`, uid, rid)
	if err != nil {
		return nil, err
	}
	ms := make([]view.MsgDisplayData, 0)
	for rows.Next() {
		var m view.MsgDisplayData
		var time time.Time
		err := rows.Scan(&m.Msg, &time, &m.Username, &m.Mine)
		if err != nil {
			return nil, err
		}
		formatted := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
			time.Year(), time.Month(), time.Day(),
			time.Hour(), time.Minute(), time.Second())
		m.Time = formatted
		ms = append(ms, m)
	}
	return ms, nil
}

func getRoomsFromUser(ctx context.Context, db *pgxpool.Pool, uid uuid.UUID) ([]*Room, error) {
	rows, err := db.Query(ctx,
		`select room.id, room.roomname, room.creator_id
            from room
            inner join room_user on room_user.room_id = room.id
            inner join "user" u on room_user.user_id = u.id
            where u.id = $1`, uid)
	if err != nil {
		return nil, err
	}
	rooms := make([]*Room, 0)
	for rows.Next() {
		room := &Room{}
		err := rows.Scan(&room.ID, &room.Name, &room.CreatorID)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// =================================== Deleting a room ===================================
// deleteRoom performs the deletion of a room and its associated entries.
//
// It deletes users associated with the room from the room_user junction table,
// then all messages related to the room from the message table, and finally
// removes the room entry from the room table. Any error encountered
// during the deletion process or transaction execution will be returned.
func deleteRoom(ctx context.Context, db *pgxpool.Pool, rid uuid.UUID) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	// We will rollback in case of an early return.
	// Since it is "deferred", it will still gets called after the function returns
	// successfully, but since the transaction will already be committed by then,
	// the rollback function will have no effect on it.
	defer tx.Rollback(ctx)
	if err := deleteAllUsersFromRoom(ctx, tx, rid); err != nil {
		return err
	}
	if err := deleteAllMessagesFromRoom(ctx, tx, rid); err != nil {
		return err
	}
	if err := deleteRoomEntry(ctx, tx, rid); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func deleteRoomEntry(ctx context.Context, tx pgx.Tx, rid uuid.UUID) error {
	_, err := tx.Exec(ctx, `delete from room where room.id = $1`, rid)
	return err
}

func deleteAllUsersFromRoom(ctx context.Context, tx pgx.Tx, rid uuid.UUID) error {
	_, err := tx.Exec(ctx, `delete from room_user ru where ru.room_id = $1`, rid)
	return err
}

func deleteAllMessagesFromRoom(ctx context.Context, tx pgx.Tx, rid uuid.UUID) error {
	_, err := tx.Exec(ctx, `delete from message where message.room_id = $1`, rid)
	return err
}

// =======================================================================================
