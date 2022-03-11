package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/navbryce/next-dorm-be/types"
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
	NumVotes              int              `db:"num_votes"`
	VoteTotal             int              `db:"vote_total"`
	CreatorId             string           `db:"creator_id"`
	CreatorDisplayName    string           `db:"display_name"`
	CommunityIdsStr       string           `db:"community_ids"`
	CommunityNamesJSONStr string           `db:"community_names"`
	Visibility            types.Visibility `db:"visibility"`
	CreatedAt             time.Time        `db:"created_at"`
	UpdatedAt             time.Time        `db:"updated_at"`
}

var postColumns = []interface{}{"p.id", "p.creator_id", "p.num_votes", "p.vote_total", "person.display_name", "p.content", "p.visibility", "p.created_at", "p.updated_at", db.Raw("JSON_ARRAYAGG(pc.community_id) as community_ids"), db.Raw("JSON_ARRAYAGG(c.name) as community_names")}

func (psdb *PlanetScaleDB) GetPostById(ctx context.Context, id int64) (*types.Post, error) {
	var post flattenedPost
	if err := psdb.sess.SQL().
		Select(postColumns...).
		From("post AS p").
		LeftJoin("post_communities as pc").On("p.id = pc.post_id").
		Join("community as c").On("pc.community_id = c.id").
		Join("person").On("p.creator_id = person.firebase_id").
		Where("p.id = ?", id).
		GroupBy("p.id", "person.firebase_id").
		IteratorContext(ctx).
		One(&post); err != nil {
		return nil, err
	}
	return buildPostFromFlattened(&post)
}

func (psdb *PlanetScaleDB) GetPosts(ctx context.Context, from *time.Time, cursor string, communityIds []int64, limit int16) ([]*types.Post, error) {
	var flattenedPosts []flattenedPost
	if err := psdb.sess.SQL().
		Select(postColumns...).
		From(
			psdb.sess.SQL().
				Select("p.id").
				From("post as p").
				LeftJoin("post_communities as pc").On("p.id=pc.post_id").
				Where("(ISNULL(?) OR (p.created_at < ? OR p.created_at = ? AND (? = '' OR p.id < ?)))", from, from, from, cursor, cursor).
				And("(ISNULL(?) OR pc.community_id IN ?)", communityIds, communityIds).
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
		NumVotes:    post.NumVotes,
		VoteTotal:   post.VoteTotal,
		Communities: communities,
		Visibility:  post.Visibility,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
	}, nil
}

func (psdb *PlanetScaleDB) GetCommunities(ctx context.Context, ids []int64) ([]*types.Community, error) {
	var where []interface{}
	if ids != nil {
		where = []interface{}{"id in ?", ids}
	}
	var communities []*types.Community
	return communities, psdb.sess.SQL().
		Select("*").
		From("community").
		Where(where...).
		IteratorContext(ctx).
		All(&communities)
}

func (psdb *PlanetScaleDB) Vote(ctx context.Context, userId string, postId int64, value int8) error {
	return psdb.sess.TxContext(ctx, func(sess db.Session) error {
		row, err := sess.SQL().QueryRowContext(ctx, `SELECT value FROM vote 
																WHERE post_id = ? AND voter_id= ?
															FOR UPDATE`,
			postId, userId)
		if err != nil {
			return err
		}
		var previousVoteValue int8
		if err := row.Scan(&previousVoteValue); err != nil {
			if err != sql.ErrNoRows {
				return err
			}
		}

		netVoteChange := value
		var numVotesChange int8
		if previousVoteValue != 0 {
			netVoteChange -= previousVoteValue

			// the previous vote value is the same as the new vote value
			if netVoteChange == 0 {
				return nil
			}

			if value == 0 {
				// delete old vote
				if _, err := sess.SQL().
					DeleteFrom("vote").
					Where("post_id = ? AND voter_id = ?", postId, userId).
					ExecContext(ctx); err != nil {
					return err
				}
				numVotesChange -= 1
			} else {
				// update existing vote
				if _, err := sess.SQL().
					Update("vote").
					Set("value", value).
					Where("post_id = ? AND voter_id = ?", postId, userId).
					ExecContext(ctx); err != nil {
					return err
				}
				numVotesChange = 0
			}
		} else if value == 0 {
			return nil
		} else {
			// insert new vote
			if _, err := sess.SQL().
				InsertInto("vote").
				Columns("voter_id", "post_id", "value").
				Values(userId, postId, value).
				ExecContext(ctx); err != nil {
				return err
			}
			numVotesChange += 1
		}

		_, err = sess.SQL().
			Update("post").
			Set("vote_total = vote_total + ?, num_votes = num_votes + ?", netVoteChange, numVotesChange).
			Where("id = ?", postId).
			ExecContext(ctx)
		return err
	}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
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
