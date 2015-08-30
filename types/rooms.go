package types

import (
	"errors"
	"fmt"
)

type RoomDescription struct {
	Visibility Visibility `json:"visibility"`
	Alias      *string    `json:"room_alias_name"`
	Name       *string    `json:"name"`
	Topic      *string    `json:"topic"`
	Invited    []UserId   `json:"invite"`
}

type Visibility int

const (
	VisibilityPrivate Visibility = 0
	VisibilityPublic  Visibility = 1
)

type JoinRule int

const (
	JoinRuleNone    JoinRule = 0
	JoinRulePublic  JoinRule = 1
	JoinRuleInvite  JoinRule = 2
	JoinRulePrivate JoinRule = 3
	JoinRuleKnock   JoinRule = 4
)

type Membership int

const (
	MembershipNone     Membership = 0
	MembershipInvited  Membership = 1
	MembershipMember   Membership = 2
	MembershipKnocking Membership = 3
	MembershipLeaving  Membership = 4
	MembershipBanned   Membership = 5
)

func (v Visibility) ToJoinRule() JoinRule {
	if v == VisibilityPublic {
		return JoinRulePublic
	} else {
		return JoinRuleInvite
	}
}

func (v *Visibility) UnmarshalJSON(bytes []byte) error {
	str := string(bytes)
	switch str {
	case "\"private\"":
		*v = VisibilityPrivate
		return nil
	case "\"public\"":
		*v = VisibilityPublic
		return nil
	}
	return errors.New("invalid visibility: " + str)
}

func (j JoinRule) ToVisibility() Visibility {
	if j == JoinRulePublic {
		return VisibilityPublic
	} else {
		return VisibilityPrivate
	}
}

func (j *JoinRule) UnmarshalJSON(bytes []byte) error {
	str := string(bytes)
	switch str {
	case "null":
		*j = JoinRuleNone
		return nil
	case "\"public\"":
		*j = JoinRulePublic
		return nil
	case "\"invite\"":
		*j = JoinRuleInvite
		return nil
	case "\"private\"":
		*j = JoinRulePrivate
		return nil
	case "\"knock\"":
		*j = JoinRuleKnock
		return nil
	}
	return errors.New("invalid join rule: " + str)
}

func (j JoinRule) String() string {
	switch j {
	case JoinRulePublic:
		return "public"
	case JoinRuleInvite:
		return "invite"
	case JoinRulePrivate:
		return "private"
	case JoinRuleKnock:
		return "knock"
	}
	return ""
}

func (j JoinRule) MarshalJSON() ([]byte, error) {
	str := j.String()
	if str == "" {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", str)), nil
}

func (m *Membership) UnmarshalJSON(bytes []byte) error {
	str := string(bytes)
	switch str {
	case "null":
		*m = MembershipNone
		return nil
	case "\"invite\"":
		*m = MembershipInvited
		return nil
	case "\"join\"":
		*m = MembershipMember
		return nil
	case "\"knock\"":
		*m = MembershipKnocking
		return nil
	case "\"leave\"":
		*m = MembershipLeaving
		return nil
	case "\"ban\"":
		*m = MembershipBanned
		return nil
	}
	return errors.New("invalid membership: " + str)
}

func (m Membership) String() string {
	switch m {
	case MembershipInvited:
		return "invite"
	case MembershipMember:
		return "join"
	case MembershipKnocking:
		return "knock"
	case MembershipLeaving:
		return "leave"
	case MembershipBanned:
		return "ban"
	}
	return ""
}

func (m Membership) MarshalJSON() ([]byte, error) {
	str := m.String()
	if str == "" {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", str)), nil
}
