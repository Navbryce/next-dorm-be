package app

import (
	"context"
	"fmt"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/model"
	"time"
)

func GetFeedForUser(
	ctx context.Context,
	db db.Database,
	user *model.User,
	cursorType PostCursorType,
	rawCursor RawCursor,
) (PostCursor, error) {
	switch cursorType {
	case PostCursorTypeMostRecent:
		// TODO: Refactor with feed interface with initialize and from existing
		if rawCursor == nil {
			return mostRecentCursorForSubbed(ctx, db, user)
		}
		return MostRecentCursorFromRaw(db, user, rawCursor)
	default:
		return nil, fmt.Errorf("unsupported feedtype %v", cursorType)
	}
}

func mostRecentCursorForSubbed(ctx context.Context, db db.Database, user *model.User) (PostCursor, error) {
	subbedIds, err := getSubbedCommunityIds(ctx, db, user)
	if err != nil {
		return nil, err
	}
	return &mostRecentCursor{
		db:          db,
		user:        user,
		Communities: subbedIds,
		LastDate:    time.Now(),
		LastId:      "",
	}, nil
}

func getSubbedCommunityIds(ctx context.Context, db db.Database, user *model.User) ([]int64, error) {
	subs, err := db.GetSubsForUser(ctx, user.Id)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(subs))
	for i, sub := range subs {
		ids[i] = sub.CommunityId
	}
	return ids, nil
}
