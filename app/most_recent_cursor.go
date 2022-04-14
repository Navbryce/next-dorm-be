package app

import (
	"context"
	"encoding/json"
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
	Communities []int64    `json:"communities,omitempty"`
	LastDate    time.Time  `json:"lastDate"`
	LastId      string     `json:"lastId"`
	ByUser      *db.ByUser `json:"byUser"`
}

// MostRecentCursorFromRaw assumes rawCursor is not nil.
// TODO Move out of app logic since parsing?
// TODO: Custom serialization logic?
func MostRecentCursorFromRaw(appDb db.Database, user *model.User, rawCursor RawCursor) (*mostRecentCursor, error) {
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

	// TODO: Make optional arguments consistent
	lastId, _ := rawCursor["lastId"].(string)

	var byUser *db.ByUser
	if byUserRaw, hasByUser := rawCursor["byUser"]; hasByUser {
		id := byUserRaw.(map[string]interface{})["id"].(string)
		byUser = &db.ByUser{Id: id}
	}

	return &mostRecentCursor{
		db:          appDb,
		user:        user,
		LastDate:    lastDate,
		LastId:      lastId,
		Communities: communities,
		ByUser:      byUser,
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

type serializableByUser struct {
	Id string `json:"id"`
}

func (mrpc *mostRecentCursor) MarshalJSON() ([]byte, error) {
	var serByUser *serializableByUser
	if mrpc.ByUser != nil {
		serByUser = &serializableByUser{mrpc.ByUser.Id}
	}
	return json.Marshal(&struct {
		Communities []int64             `json:"communities,omitempty"`
		LastDate    time.Time           `json:"lastDate"`
		LastId      string              `json:"lastId"`
		ByUser      *serializableByUser `json:"byUser,omitempty"`
	}{
		Communities: mrpc.Communities,
		LastDate:    mrpc.LastDate,
		LastId:      mrpc.LastId,
		ByUser:      serByUser,
	})
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
		ByUser:       mrpc.ByUser,
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
		ByUser:      mrpc.ByUser,
	}
}
