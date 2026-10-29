package telegram

// Update represents a Telegram Update object.
type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

// Message represents a Telegram Message object.
type Message struct {
	MessageID int         `json:"message_id"`
	From      *User       `json:"from"`
	Chat      *Chat       `json:"chat"`
	Text      string      `json:"text"`
	Photo     []PhotoSize `json:"photo"`
}

// User represents a Telegram User.
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

// Chat represents a Telegram Chat.
type Chat struct {
	ID int64 `json:"id"`
}

// PhotoSize represents one size of a photo or a file / sticker.
type PhotoSize struct {
	FileID   string `json:"file_id"`
	FileSize int    `json:"file_size"`
}
