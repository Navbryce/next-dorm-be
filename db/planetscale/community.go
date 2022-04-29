package planetscale

import (
	"context"
	appDb "github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/upper/db/v4"
)

type CommunityDB struct {
	sess db.Session
}

func getCommunityDb(sess db.Session) *CommunityDB {
	return &CommunityDB{sess}
}

func (cdb *CommunityDB) CreateCommunity(ctx context.Context, name string) (int64, error) {
	res, err := cdb.sess.SQL().
		InsertInto("community").
		Values(name).
		Columns("name").
		ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetCommunitiesByIds gets communities. nil ids gets all communities
func (cdb *CommunityDB) GetCommunitiesByIds(ctx context.Context, ids []int64, opts *appDb.GetCommunitiesQueryOpts) ([]*model.CommunityWithSubStatus, error) {
	var where []interface{}
	if ids != nil {
		where = []interface{}{"id in ?", ids}
	}
	var communities []*model.CommunityWithSubStatus
	if err := cdb.sess.SQL().
		Select("c.id", "c.name", "c.created_at", db.Raw("s.user_id IS NOT NULL AS is_subscribed")).
		From("community as c").
		// TODO: Change to only join if user id is provided
		LeftJoin("subscription as s").On("c.id = s.community_id AND s.user_id = ?", opts.ForUserId).
		Where(where...).
		IteratorContext(ctx).
		All(&communities); err != nil {
		return nil, err
	}
	return communities, nil
}
