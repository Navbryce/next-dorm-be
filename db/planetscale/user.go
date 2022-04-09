package planetscale

import (
	"context"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/upper/db/v4"
)

type UserDB struct {
	sess db.Session
}

func getUserDB(sess db.Session) *UserDB {
	return &UserDB{sess}
}

func (udb *UserDB) CreateUser(ctx context.Context, user *model.User) error {
	_, err := udb.sess.Collection("person").
		Insert(user)
	return err
}

func (udb *UserDB) GetUser(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	if err := udb.sess.SQL().
		Select("*").
		From("person").
		Where("firebase_id = ?", id).
		IteratorContext(ctx).
		One(&user); err != nil {
		if err == db.ErrNoMoreRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
