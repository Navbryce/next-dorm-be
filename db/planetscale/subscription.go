package planetscale

import (
	"context"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/upper/db/v4"
)

type SubscriptionDB struct {
	sess db.Session
}

func getSubscriptionDB(sess db.Session) *SubscriptionDB {
	return &SubscriptionDB{sess}
}

func (udb *UserDB) CreateSubForUser(ctx context.Context, sub *model.Subscription) error {
	_, err := udb.sess.Collection("subscription").
		Insert(sub)
	return err
}

func (udb *UserDB) DeleteSubForUser(ctx context.Context, sub *model.Subscription) error {
	return udb.sess.WithContext(ctx).
		Collection("subscription").
		Find("user_id = ? AND community_id = ?", sub.UserId, sub.CommunityId).
		Delete()
}

func (udb *UserDB) GetSubsForUser(ctx context.Context, userId string) ([]*model.Subscription, error) {
	var subs []*model.Subscription
	err := udb.sess.WithContext(ctx).
		Collection("subscription").
		Find("user_id = ?", userId).
		All(&subs)
	return subs, err
}
