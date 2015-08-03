package types

import "errors"

type RoomDescription struct {
	Visibility Visibility `json:"visibility"`
	Alias      *Alias     `json:"room_alias_name"`
	Name       *string    `json:"name"`
	Topic      *string    `json:"topic"`
	Invited    []UserId   `json:"invite"`
}

type Visibility int

const (
	VisibilityPrivate Visibility = 0
	VisibilityPublic             = 1
)

type JoinRule int

const (
	JoinRuleNone   JoinRule = 0
	JoinRulePublic          = 1
	JoinRuleInvite          = 2
)

type Membership int

const (
	MembershipNone     Membership = 0
	MembershipInvited             = 1
	MembershipMember              = 2
	MembershipKnocking            = 3
	MembershipLeaving             = 4
	MembershipBanned              = 5
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
	return errors.New("invalid visibility identifier: '" + str + "'")
}

func (v Visibility) MarshalJSON() ([]byte, error) {
	switch v {
	case VisibilityPrivate:
		return []byte("\"private\""), nil
	case VisibilityPublic:
		return []byte("\"public\""), nil
	}
	return nil, errors.New("invalid visibility value: '" + string(v) + "'")
}

func (v *JoinRule) UnmarshalJSON(bytes []byte) error {
	str := string(bytes)
	switch str {
	case "\"public\"":
		*v = JoinRulePublic
		return nil
	case "\"invite\"":
		*v = JoinRuleInvite
		return nil
	}
	return errors.New("invalid join rule identifier: '" + str + "'")
}

func (v JoinRule) MarshalJSON() ([]byte, error) {
	switch v {
	case JoinRuleNone:
		return []byte("null"), nil
	case JoinRulePublic:
		return []byte("\"public\""), nil
	case JoinRuleInvite:
		return []byte("\"invite\""), nil
	}
	return nil, errors.New("invalid join rule value: '" + string(v) + "'")
}

func (v *Membership) UnmarshalJSON(bytes []byte) error {
	str := string(bytes)
	switch str {
	case "null":
		*v = MembershipNone
		return nil
	case "\"invite\"":
		*v = MembershipInvited
		return nil
	case "\"join\"":
		*v = MembershipMember
		return nil
	case "\"knock\"":
		*v = MembershipKnocking
		return nil
	case "\"leave\"":
		*v = MembershipLeaving
		return nil
	case "\"ban\"":
		*v = MembershipBanned
		return nil
	}
	return errors.New("invalid membership identifier: '" + str + "'")
}

func (v Membership) MarshalJSON() ([]byte, error) {
	switch v {
	case MembershipNone:
		return []byte("null"), nil
	case MembershipInvited:
		return []byte("\"invite\""), nil
	case MembershipMember:
		return []byte("\"join\""), nil
	case MembershipKnocking:
		return []byte("\"knock\""), nil
	case MembershipLeaving:
		return []byte("\"leave\""), nil
	case MembershipBanned:
		return []byte("\"ban\""), nil
	}
	return nil, errors.New("invalid membership value: '" + string(v) + "'")
}
