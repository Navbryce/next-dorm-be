package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/heroku/go-getting-started/types"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"os"
	"time"
)

type PlanetScaleDB struct {
	sess  db.Session
	sqlDB *sql.DB
}

func GetDatabase() (Database, error) {
	// TODO: MOVE CONFIG PARSING AND VALIDATINO TO SEPARATE MODULE
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/next-dorm?tls=true&parseTime=true",
			os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST")))
	if err != nil {
		return nil, err
	}

	sess, err := mysql.New(db)
	if err != nil {
		return nil, err
	}

	return &PlanetScaleDB{
		sess:  sess,
		sqlDB: db,
	}, nil
}

func (psdb *PlanetScaleDB) GetSQLDB() *sql.DB {
	return psdb.sqlDB
}

func (psdb *PlanetScaleDB) Close() error {
	return psdb.sess.Close()
}

func (psdb *PlanetScaleDB) CreateCommunity(ctx context.Context, name string) (int64, error) {
	res, err := psdb.sess.SQL().
		InsertInto("community").
		Values(name).
		Columns("name").
		ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (psdb *PlanetScaleDB) CreatePost(ctx context.Context, creatorId string, post *CreatePost) (int64, error) {
	var postId int64
	err := psdb.sess.TxContext(ctx, func(sess db.Session) error {
		res, err := sess.SQL().
			InsertInto("post").
			Columns("creator_id", "content", "visibility").
			Values(creatorId, post.Content, post.Visibility).
			ExecContext(ctx)
		if err != nil {
			return err
		}
		postId, err = res.LastInsertId()
		if err != nil {
			return err
		}

		batchInserter := sess.SQL().
			InsertInto("post_communities").
			Columns("post_id", "community_id").
			Batch(len(post.Communities))
		for _, communityId := range post.Communities {
			batchInserter.Values(postId, communityId)
		}
		batchInserter.Done()
		return batchInserter.Wait()
	}, nil)
	return postId, err
}

type flattenedPost struct {
	Id                    int64            `db:"id"`
	Content               string           `db:"content"`
	CreatorId             string           `db:"creator_id"`
	CreatorDisplayName    string           `db:"display_name"`
	CommunityIdsStr       string           `db:"community_ids"`
	CommunityNamesJSONStr string           `db:"community_names"`
	Visibility            types.Visibility `db:"visibility"`
	CreatedAt             time.Time        `db:"created_at"`
	UpdatedAt             time.Time        `db:"updated_at"`
}

func (psdb *PlanetScaleDB) GetPostById(ctx context.Context, id int64) (*types.Post, error) {
	var post flattenedPost
	if err := psdb.sess.SQL().
		Select("p.id", "p.content", "p.visibility", "p.created_at", "p.updated_at", db.Raw("JSON_ARRAYAGG(pc.community_id) as community_ids"), db.Raw("JSON_ARRAYAGG(c.name) as community_names")).
		From("post AS p").
		LeftJoin("post_communities as pc").On("p.id = pc.post_id").
		Join("community as c").On("pc.community_id = c.id").
		Where("p.id = ?", id).
		GroupBy("p.id").
		IteratorContext(ctx).
		One(&post); err != nil {
		return nil, err
	}
	return buildPostFromFlattened(&post)
}

func (psdb *PlanetScaleDB) GetPosts(ctx context.Context, from *time.Time, cursor string, communityIds []int64, limit int16) ([]*types.Post, error) {
	var flattenedPosts []flattenedPost
	if err := psdb.sess.SQL().
		Select("p.id", "p.creator_id", "person.display_name", "p.content", "p.visibility", "p.created_at", "p.updated_at", db.Raw("JSON_ARRAYAGG(pc.community_id) as community_ids"), db.Raw("JSON_ARRAYAGG(c.name) as community_names")).
		From(
			psdb.sess.SQL().
				Select("p.id").
				From("post as p").
				LeftJoin("post_communities as pc").On("p.id=pc.post_id").
				Where("(ISNULL(?) OR (p.created_at < ? OR p.created_at = ? AND p.id < ?))", from, from, from, cursor).
				And("pc.community_id IN ?", communityIds).
				GroupBy("p.id")).
		As("p_ids").
		Join("post as p").On("p_ids.id = p.id").
		LeftJoin("post_communities as pc").On("p.id = pc.post_id").
		Join("community as c").On("pc.community_id = c.id").
		Join("person").On("p.creator_id = person.firebase_id").
		OrderBy("p.created_at DESC", "p.id DESC").
		GroupBy("p.id", "person.firebase_id").
		Limit(int(limit)).
		IteratorContext(ctx).
		All(&flattenedPosts); err != nil {
		return nil, err
	}
	posts := make([]*types.Post, len(flattenedPosts))
	for i, flattened := range flattenedPosts {
		post, err := buildPostFromFlattened(&flattened)
		if err != nil {
			return nil, err
		}
		posts[i] = post
	}
	return posts, nil
}

func buildPostFromFlattened(post *flattenedPost) (*types.Post, error) {
	var communityIds []int64
	if err := json.Unmarshal([]byte(post.CommunityIdsStr), &communityIds); err != nil {
		return nil, err
	}

	var communityNames []string
	if err := json.Unmarshal([]byte(post.CommunityNamesJSONStr), &communityNames); err != nil {
		return nil, err
	}

	communities := make([]*types.Community, len(communityIds))
	for i, communityId := range communityIds {
		communities[i] = &types.Community{
			Id:   communityId,
			Name: communityNames[i],
		}
	}

	return &types.Post{
		Id: post.Id,
		Creator: &types.DisplayableUser{
			Id:          post.CreatorId,
			DisplayName: post.CreatorDisplayName,
		},
		Content:     post.Content,
		Communities: communities,
		Visibility:  post.Visibility,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
	}, nil
}

func (psdb *PlanetScaleDB) GetCommunities(ctx context.Context) ([]*types.Community, error) {
	var communities []*types.Community
	return communities, psdb.sess.SQL().
		Select("*").
		From("community").
		IteratorContext(ctx).
		All(communities)
}

func (psdb *PlanetScaleDB) CreateUser(ctx context.Context, user *types.User) error {
	_, err := psdb.sess.Collection("person").
		Insert(user)
	return err
}

func (psdb *PlanetScaleDB) GetUser(ctx context.Context, id string) (*types.User, error) {
	var user types.User
	if err := psdb.sess.SQL().
		Select("*").
		From("person").
		Where("firebase_id = ?", id).
		IteratorContext(ctx).
		One(&user); err != nil {
		return nil, err
	}
	return &user, nil

}
