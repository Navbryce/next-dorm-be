package app

import (
	"context"
	appDb "github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/model"
	"strconv"
	"time"
)

type LastVoteTotal struct {
	Val int64 `json:"val"`
}

func (lvt *LastVoteTotal) ToDBFilter() *appDb.IntFilter {
	if lvt == nil {
		return nil
	}
	return &appDb.IntFilter{Val: lvt.Val}
}

type Since string

func (s *Since) ToTime() *time.Time {
	switch *s {
	case SinceToday:
		val := time.Now().Add(-24 * time.Hour)
		return &val
	default:
		panic("not defined for since value")
	}
}

const (
	SinceToday = "TODAY"
)

type MostPopularCursor struct {
	Communities   []int64             `json:"communities,omitempty"`
	Since         *Since              `json:"since,omitempty"`
	LastVoteTotal *LastVoteTotal      `json:"lastVoteTotal,omitempty"` // TODO: Will have to change to nilable pointer
	LastId        string              `json:"lastId"`
	ByUser        *SerializableByUser `json:"byUser,omitempty"`
	Visibility    *model.Visibility   `json:"visibility,omitempty"`
}

// TODO: Split into filter params and persisted cursor params
func (mpc *MostPopularCursor) Posts(ctx context.Context, db appDb.Database, user *model.LocalUser, cursorOpts *PostCursorOpts) (posts []*model.Post, cursor interface{}, err error) {
	// TODO: PERMS CHECKS?
	voteHistoryOf := ""
	if user != nil {
		voteHistoryOf = user.Id
	}

	var byUser *appDb.ByUser
	if mpc.ByUser != nil {
		byUser = &appDb.ByUser{Id: mpc.ByUser.Id}
	}
	var since *time.Time
	if mpc.Since != nil {
		since = mpc.Since.ToTime()
	}

	// TODO: Create specific query for paged by vote total
	posts, err = db.GetPosts(ctx, &appDb.PostsListQuery{
		CommunityIds: mpc.Communities,
		ByUser:       byUser,
		Visibility:   mpc.Visibility,
		PageByVote: &appDb.ByVotePaging{
			MaxUpvotes: mpc.LastVoteTotal.ToDBFilter(),
			LastId:     mpc.LastId,
			Since:      since,
		},
		PostsListQueryOpts: &appDb.PostsListQueryOpts{
			Limit:         cursorOpts.Limit,
			VoteHistoryOf: voteHistoryOf,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return posts, mpc.buildCursorForNextPage(posts), nil
}

func (mpc *MostPopularCursor) buildCursorForNextPage(previousPosts []*model.Post) *MostPopularCursor {
	if len(previousPosts) == 0 {
		return nil
	}
	return &MostPopularCursor{
		Communities:   mpc.Communities,
		Since:         mpc.Since,
		LastId:        strconv.FormatInt(previousPosts[len(previousPosts)-1].Id, 10),
		ByUser:        mpc.ByUser,
		LastVoteTotal: &LastVoteTotal{previousPosts[len(previousPosts)-1].VoteTotal},
	}
}

func (mpc *MostPopularCursor) WithCommunities(communities []int64) *MostPopularCursor {
	newCursor := *mpc
	newCursor.Communities = communities
	return &newCursor
}
