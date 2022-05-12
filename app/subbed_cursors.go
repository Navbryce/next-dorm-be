package app

import (
	"context"
	"errors"
	appDb "github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/model"
)

// TODO: Delete subbed cursors and add subbedOnly param to normal cursors
type SubbedMostRecentCursor struct {
	MostRecentCursor
}

func (s *SubbedMostRecentCursor) Posts(ctx context.Context, db appDb.Database, user *model.User, cursorOpts *PostCursorOpts) (posts []*model.Post, cursor interface{}, err error) {
	if s != nil && s.Communities != nil {
		return s.MostRecentCursor.Posts(ctx, db, user, cursorOpts)
	}
	communities, err := fetchSubbedCommunityIds(ctx, db, user)
	if err != nil {
		return nil, nil, err
	}

	return s.WithCommunities(communities).Posts(ctx, db, user, cursorOpts)
}

type SubbedMostPopularCursor struct {
	MostPopularCursor
}

func (s *SubbedMostPopularCursor) Posts(ctx context.Context, db appDb.Database, user *model.User, cursorOpts *PostCursorOpts) (posts []*model.Post, cursor interface{}, err error) {
	if s != nil && s.Communities != nil {
		return s.MostPopularCursor.Posts(ctx, db, user, cursorOpts)
	}
	communities, err := fetchSubbedCommunityIds(ctx, db, user)
	if err != nil {
		return nil, nil, err
	}

	return s.WithCommunities(communities).Posts(ctx, db, user, cursorOpts)
}

func fetchSubbedCommunityIds(ctx context.Context, db appDb.Database, user *model.User) ([]int64, error) {
	if user == nil {
		return nil, errors.New("must be logged in to fetch subs")
	}
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
