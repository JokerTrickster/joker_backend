package request

type CreateRoomRequest struct {
	RoomType string `json:"room_type" validate:"required,oneof=random private"`
}

type JoinRoomRequest struct {
	RoomCode string `json:"room_code" validate:"required,len=4"`
}

type SetReadyRequest struct {
	IsReady bool `json:"is_ready"`
}
