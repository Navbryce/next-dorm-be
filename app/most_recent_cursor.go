package app

import (
	"context"
	appDb "github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/model"
	"strconv"
	"time"
)

type MostRecentCursor struct {
	Communities []int64             `json:"communities,omitempty"`
	LastDate    *time.Time          `json:"lastDate,omitempty"`
	LastId      string              `json:"lastId"`
	ByUser      *SerializableByUser `json:"byUser,omitempty"`
	Visibility  *model.Visibility   `json:"visibility,omitempty"`
}

type SerializableByUser struct {
	Id string `json:"id"`
}

func (mrpc *MostRecentCursor) Posts(ctx context.Context, db appDb.Database, user *model.User, cursorOpts *PostCursorOpts) (posts []*model.Post, cursor interface{}, err error) {
	// TODO: PERMS CHECKS?
	voteHistoryOf := ""
	if user != nil {
		voteHistoryOf = user.Id
	}

	var byUser *appDb.ByUser
	if mrpc.ByUser != nil {
		byUser = &appDb.ByUser{Id: mrpc.ByUser.Id}
	}

	posts, err = db.GetPosts(ctx, &appDb.PostsListQuery{
		CommunityIds: mrpc.Communities,
		ByUser:       byUser,
		Visibility:   mrpc.Visibility,
		PageByDate: &appDb.ByDatePaging{
			From:   mrpc.LastDate,
			LastId: mrpc.LastId,
		},
		PostsListQueryOpts: &appDb.PostsListQueryOpts{
			Limit:         cursorOpts.Limit,
			VoteHistoryOf: voteHistoryOf,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return posts, mrpc.buildCursorForNextPage(posts), nil
}

func (mrpc *MostRecentCursor) buildCursorForNextPage(previousPosts []*model.Post) *MostRecentCursor {
	if len(previousPosts) == 0 {
		return nil
	}
	return &MostRecentCursor{
		Communities: mrpc.Communities,
		LastDate:    &previousPosts[len(previousPosts)-1].CreatedAt,
		LastId:      strconv.FormatInt(previousPosts[len(previousPosts)-1].Id, 10),
		ByUser:      mrpc.ByUser,
	}
}

func (mrpc *MostRecentCursor) WithCommunities(communities []int64) *MostRecentCursor {
	newCursor := *mrpc
	newCursor.Communities = communities
	return &newCursor
}
