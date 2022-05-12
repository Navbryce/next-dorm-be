package app

import (
	"encoding/json"
	"errors"
)

const (
	PostCursorTypeMostRecent        PostCursorType = "MOST_RECENT"
	PostCursorTypeSubbedMostRecent  PostCursorType = "SUBBED_MOST_RECENT"
	PostCursorTypeMostPopular       PostCursorType = "MOST_POPULAR"
	PostCursorTypeSubbedMostPopular PostCursorType = "SUBBED_MOST_POPULAR"
)

var UnknownCursorTypeErr = errors.New("unknown cursor type")

type TaggedUnionCursor struct {
	PostCursor
	CursorType PostCursorType // TODO: Field not really used. Keep?
}

func (tuc *TaggedUnionCursor) UnmarshalJSON(data []byte) error {
	if tuc == nil {
		return nil
	}
	var rawJsonWithType struct {
		CursorType PostCursorType   `json:"cursorType"`
		Raw        *json.RawMessage `json:"cursor"`
	}
	if err := json.Unmarshal(data, &rawJsonWithType); err != nil {
		return err
	}

	tuc.CursorType = rawJsonWithType.CursorType

	var cursorRef interface{}
	switch rawJsonWithType.CursorType {
	case PostCursorTypeMostRecent:
		cursorRef = &MostRecentCursor{}
	case PostCursorTypeSubbedMostRecent:
		cursorRef = &SubbedMostRecentCursor{}
	case PostCursorTypeMostPopular:
		cursorRef = &MostPopularCursor{}
	case PostCursorTypeSubbedMostPopular:
		cursorRef = &SubbedMostPopularCursor{}
	default:
		return UnknownCursorTypeErr
	}

	if rawJsonWithType.Raw != nil {
		if err := json.Unmarshal(*rawJsonWithType.Raw, cursorRef); err != nil {
			return err
		}
	}

	tuc.PostCursor = cursorRef.(PostCursor)
	return nil
}

func (tuc *TaggedUnionCursor) MarshalJSON() ([]byte, error) {
	panic("should not be unmarshalled")
}
