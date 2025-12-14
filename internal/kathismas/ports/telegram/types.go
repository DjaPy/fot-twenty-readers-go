package telegram

import "github.com/gofrs/uuid/v5"

type RegistrationState string

const (
	StateIdle                 RegistrationState = "idle"
	StateAwaitingName         RegistrationState = "awaiting_name"
	StateAwaitingGroup        RegistrationState = "awaiting_group"
	StateAwaitingReaderNumber RegistrationState = "awaiting_reader_number"
	StateAwaitingConfirm      RegistrationState = "awaiting_confirm"
)

type UserSession struct {
	State        RegistrationState
	Username     string
	GroupID      uuid.UUID
	GroupName    string
	ReaderNumber int8
}
