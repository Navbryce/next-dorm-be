package app

import (
	"context"
	"fmt"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/navbryce/next-dorm-be/util"
	"strconv"
	"time"
)

type mostRecentCursor struct {
	db          db.Database
	user        *model.User
	Communities []int64   `json:"communities,omitempty"`
	LastDate    time.Time `json:"lastDate"`
	LastId      string    `json:"lastId"`
}

// MostRecentCursorFromRaw assumes rawCursor is not nil.
// TODO Move out of app logic since parsing?
func MostRecentCursorFromRaw(db db.Database, user *model.User, rawCursor RawCursor) (*mostRecentCursor, error) {
	lastDateStr, hasLastDate := rawCursor["lastDate"].(string)
	lastDate := time.Now()
	if hasLastDate {
		var err error
		lastDate, err = util.ParseTime(lastDateStr)
		if err != nil {
			return nil, err
		}
	}

	rawCommunities, hasCommunities := rawCursor["communities"]
	var communities []int64
	if hasCommunities && rawCommunities != nil {
		var err error
		communities, err = castCommunitiesFromCursor(rawCommunities)
		if err != nil {
			return nil, err
		}
	}

	lastId, _ := rawCursor["lastId"].(string)

	return &mostRecentCursor{
		db:          db,
		user:        user,
		LastDate:    lastDate,
		LastId:      lastId,
		Communities: communities,
	}, nil
}

// TODO: Automate in some sort of req validation?
func castCommunitiesFromCursor(raw interface{}) ([]int64, error) {
	rawArray := raw.([]interface{})
	communities := make([]int64, len(rawArray))
	for i, rawId := range rawArray {
		floatId, isOk := rawId.(float64)
		if !isOk {
			return nil, fmt.Errorf("id %v is of wrong format", rawArray[i])
		}
		communities[i] = int64(floatId)
	}
	return communities, nil
}

func (mrpc *mostRecentCursor) Posts(ctx context.Context, cursorOpts *PostCursorOpts) (posts []*model.Post, cursor interface{}, err error) {
	voteHistoryOf := ""
	if mrpc.user != nil {
		voteHistoryOf = mrpc.user.Id
	}

	posts, err = mrpc.db.GetPosts(ctx, &db.PostsListQuery{
		From:         &mrpc.LastDate,
		LastId:       mrpc.LastId,
		CommunityIds: mrpc.Communities,
		PostsListQueryOpts: &db.PostsListQueryOpts{
			// TODO: ADD configurable Limit for cursor (cursor opts in a separate struct?)
			Limit:         cursorOpts.Limit,
			VoteHistoryOf: voteHistoryOf,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return posts, mrpc.buildCursorForNextPage(posts), nil
}

func (mrpc *mostRecentCursor) buildCursorForNextPage(previousPosts []*model.Post) *mostRecentCursor {
	if len(previousPosts) == 0 {
		return nil
	}
	return &mostRecentCursor{
		db:          mrpc.db,
		user:        mrpc.user,
		Communities: mrpc.Communities,
		LastDate:    previousPosts[len(previousPosts)-1].CreatedAt,
		LastId:      strconv.FormatInt(previousPosts[len(previousPosts)-1].Id, 10),
	}
}
