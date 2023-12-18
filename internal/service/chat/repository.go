package chat

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Room struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"roomname"`
	CreatorID uuid.UUID `json:"creator_id"`
}

type RoomUser struct {
	Rid uuid.UUID `json:"room_id"`
	Uid uuid.UUID `json:"user_id"`
}

func addRoom(ctx context.Context, db *pgxpool.Pool, r *Room) error {
	_, err := db.Exec(ctx, `insert into room(id, roomname, creator_id) values($1, $2, $3)`, r.ID, r.Name, r.CreatorID)
	return err
}

func getRoomByID(ctx context.Context, db *pgxpool.Pool, id uuid.UUID) (*Room, error) {
	r := &Room{}
	err := db.QueryRow(ctx, `select * from room where room.id = $1`, id).Scan(&r.ID, &r.Name, &r.CreatorID)
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

func addMembership(ctx context.Context, db *pgxpool.Pool, ru *RoomUser) error {
	_, err := db.Exec(ctx, `insert into room_user(room_id, user_id) values($1, $2)`, ru.Rid, ru.Uid)
	return err
}

func isAMember(ctx context.Context, db *pgxpool.Pool, ru *RoomUser) (bool, error) {
	exists := false
	err := db.QueryRow(ctx, `select exists(select 1 from room_user ru where ru.room_id = $1 and ru.user_id = $2)`, ru.Rid, ru.Uid).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// getUserRooms - given uid, return all rooms that the uid is apart of
func getRidsForUser(ctx context.Context, db *pgxpool.Pool, uid uuid.UUID) ([]uuid.UUID, error) {
	rows, err := db.Query(ctx, `select (room_id) from room_user ru where ru.user_id = $1`, uid)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	rids := make([]uuid.UUID, 0)
	for rows.Next() {
		var rid uuid.UUID
		err := rows.Scan(&rid)
		if err != nil {
			return nil, err
		}
		rids = append(rids, rid)
	}
	return rids, nil
}

func getUidsFromRoom(ctx context.Context, db *pgxpool.Pool, rid uuid.UUID) ([]uuid.UUID, error) {
	rows, err := db.Query(ctx, `select (user_id) from room_user ru where ru.room_id = $1`, rid)
	if err != nil {
		return nil, err
	}
	uids := make([]uuid.UUID, 0)
	for rows.Next() {
		var uid uuid.UUID
		err := rows.Scan(&uid)
		if err != nil {
			return nil, err
		}
		uids = append(uids, uid)
	}
	return uids, nil
}

func deleteRoomById(ctx context.Context, db *pgxpool.Pool, rid uuid.UUID) error {
	_, err := db.Exec(ctx, `delete from room where room.id = $1`, rid)
	return err
}

func removeMembership(ctx context.Context, db *pgxpool.Pool, ru *RoomUser) error {
	_, err := db.Exec(ctx, `delete from room_user ru where ru.room_id = $1 and ru.user_id = $2`, ru.Rid, ru.Uid)
	return err
}
