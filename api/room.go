package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RoomUser struct {
	Order int
	UUID  uuid.UUID
	State string
}

type Room struct {
	Code     string
	Owner    uuid.UUID
	Users    []*RoomUser
	Created  time.Time
	Capacity int
}

const RoomCodeLen = 6

var rooms map[string]*Room = make(map[string]*Room)

func createRoom(owner *User) *Room {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, RoomCodeLen)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	room := &Room{
		Code:     string(b),
		Owner:    owner.UUID,
		Users:    []*RoomUser{},
		Created:  time.Now(),
		Capacity: 2,
	}
	rooms[room.Code] = room
	room.joinUser(owner)
	return room
}

func (room *Room) joinUser(user *User) error {
	if len(room.Users) == room.Capacity {
		return fmt.Errorf("The room is full")
	}
	user.Room = room.Code
	room.Users = append(room.Users, &RoomUser{
		len(room.Users),
		user.UUID,
		"stopped",
	})
	return nil
}

func handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)
	room := createRoom(user)
	user.updateCookie(w, r)
	JSONResponse(w, 200, room)
}

func handleJoinRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)
	room, ok := rooms[mux.Vars(r)["code"]]
	if !ok {
		JSONError(w, 400, "That room does not exist")
		return
	}
	if err := room.joinUser(user); err != nil {
		JSONError(w, 400, err.Error())
	}
	JSONResponse(w, 200, room)
}
