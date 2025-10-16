package recordTypes

import (
	"time"
)

// User holds the contents of a row from the adm_users table.
type User struct {
	ID                    int32     `json:"usr_id"`
	UUID                  string    `json:"usr_uuid"`
	LoginName             string    `json:"usr_login_name"`
	Password              string    `json:"usr_password"`
	Photo                 []byte    `json:"usr_photo"`
	Text                  string    `json:"usr_text"`
	PasswordResetID       string    `json:"usr_pw_reset_id"`
	PasswordRestTimestamp time.Time `json:"usr_pw_reset_timestamp"`
	LastLogin             time.Time `json:"usr_last_login"`
	ActualLogin           time.Time `json:"usr_actual_login"`
	NumberLogin           int       `json:"usr_number_login"`
	DateInvalid           time.Time `json:"usr_date_invalid"`
	NumberInvalid         int       `json:"usr_number_invalid"`
	IDCreate              int       `json:"usr_usr_id_create"`
	TimeStampCreate       time.Time `json:"usr_timestamp_create"`
	IDChange              int       `json:"usr_usr_id_change"`
	TimestampChange       time.Time `json:"usr_timestamp_change"`
	Valid                 bool      `json:"usr_valid"`
}

// Role holds the data about a role in the adm_roles table.
// These are pre-populated.  We only need the rol_id and the
// rol_name field.
type Role struct {
	ID   int64  `json:"rol_id"`
	Name string `json:"rol_name"`
}

// Member holds data from an adm_members record.  There are
// many members for each user, one per role (Member, Admin etc)
type Member struct {
	ID        int64  `json:"mem_id"`
	UserID    int64  `json:"mem_usr_id"`
	RoleID    int64  `json:"mem_rol_id"`
	UUID      string `json:"mem_uuid"`
	StartDate string `json:"mem_begin"`
	EndDate   string `json:"mem_end"`
}
