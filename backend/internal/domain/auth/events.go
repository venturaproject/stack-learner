package auth

import "boilerplate-learner/internal/domain/shared"

type UserLoggedIn struct {
	shared.BaseEvent
	UserID int32
	Email  string
}

func NewUserLoggedIn(id int32, email string) UserLoggedIn {
	return UserLoggedIn{BaseEvent: shared.NewBaseEvent(), UserID: id, Email: email}
}
func (UserLoggedIn) EventName() string { return "user.logged_in" }

type UserLoggedOut struct {
	shared.BaseEvent
	UserID int32
}

func NewUserLoggedOut(id int32) UserLoggedOut {
	return UserLoggedOut{BaseEvent: shared.NewBaseEvent(), UserID: id}
}
func (UserLoggedOut) EventName() string { return "user.logged_out" }

type UserCreated struct {
	shared.BaseEvent
	UserID int32
	Email  string
}

func NewUserCreated(id int32, email string) UserCreated {
	return UserCreated{BaseEvent: shared.NewBaseEvent(), UserID: id, Email: email}
}
func (UserCreated) EventName() string { return "user.created" }

type UserUpdated struct {
	shared.BaseEvent
	UserID int32
}

func NewUserUpdated(id int32) UserUpdated {
	return UserUpdated{BaseEvent: shared.NewBaseEvent(), UserID: id}
}
func (UserUpdated) EventName() string { return "user.updated" }

type UserDeleted struct {
	shared.BaseEvent
	UserID int32
}

func NewUserDeleted(id int32) UserDeleted {
	return UserDeleted{BaseEvent: shared.NewBaseEvent(), UserID: id}
}
func (UserDeleted) EventName() string { return "user.deleted" }
