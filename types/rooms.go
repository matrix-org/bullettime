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
