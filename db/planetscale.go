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

func (psdb *PlanetScaleDB) CreatePost(ctx context.Context, post *CreatePost) (int64, error) {
	var postId int64
	err := psdb.sess.TxContext(ctx, func(sess db.Session) error {
		metadataId, err := insertContentMetadata(ctx, sess, post.CreateContentMetadata)
		if err != nil {
			return err
		}
		res, err := sess.SQL().
			InsertInto("post").
			Columns("title", "content", "metadata_id").
			Values(post.Title, post.Content, metadataId).
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

func (psdb *PlanetScaleDB) CreateComment(ctx context.Context, req *CreateComment) (int64, error) {
	var commentId int64
	err := psdb.sess.TxContext(ctx, func(sess db.Session) error {
		metadataId, err := insertContentMetadata(ctx, sess, req.CreateContentMetadata)
		if err != nil {
			return err
		}

		res, err := sess.SQL().
			InsertInto("comment").
			Values(metadataId, req.ParentMetadataId, req.Content).
			Columns("metadata_id", "parent_metadata_id", "content").
			ExecContext(ctx)
		if err != nil {
			return err
		}
		commentId, err = res.LastInsertId()
		return err
	}, &sql.TxOptions{})
	return commentId, err
}

func insertContentMetadata(ctx context.Context, sess db.Session, metadata *CreateContentMetadata) (id int64, err error) {
	res, err := sess.SQL().
		InsertInto("content_metadata").
		Columns("creator_id", "creator_alias", "visibility").
		Values(metadata.CreatorId, metadata.CreatorAlias, metadata.Visibility).
		ExecContext(ctx)
	if err != nil {
		return 0, nil
	}
	return res.LastInsertId()
}

type flattenedMetadata struct {
	Id                 int64            `db:"id"`
	NumVotes           int              `db:"num_votes"`
	VoteTotal          int              `db:"vote_total"`
	CreatorId          string           `db:"creator_id"`
	CreatorDisplayName string           `db:"display_name"`
	CreatorAlias       string           `db:"creator_alias"`
	Visibility         types.Visibility `db:"visibility"`
	CreatedAt          time.Time        `db:"created_at"`
	UpdatedAt          time.Time        `db:"updated_at"`
}

type flattenedPost struct {
	flattenedMetadata     `db:",inline"`
	Id                    int64  `db:"id"`
	Title                 string `db:"title"`
	Content               string `db:"content"`
	CommunityIdsStr       string `db:"community_ids"`
	CommunityNamesJSONStr string `db:"community_names"`
}

var contentMetadataColumns = []interface{}{
	"cm.creator_id",
	"person.display_name",
	"cm.creator_alias",
	"cm.num_votes",
	"cm.vote_total",
	"cm.visibility",
	"cm.created_at",
	"cm.updated_at",
}

var postColumns = append([]interface{}{
	"p.id",
	"p.title",
	"p.content",
	db.Raw("JSON_ARRAYAGG(pc.community_id) as community_ids"), db.Raw("JSON_ARRAYAGG(c.name) as community_names"),
}, contentMetadataColumns...)

func (psdb *PlanetScaleDB) GetPostById(ctx context.Context, id int64) (*types.Post, error) {
	var post flattenedPost
	if err := psdb.sess.SQL().
		Select(postColumns...).
		From("post AS p").
		Join("content_metadata as cm").On("p.metadata_id = cm.id").
		LeftJoin("post_communities as pc").On("p.id = pc.post_id").
		Join("person").On("cm.creator_id = person.firebase_id").
		Join("community as c").On("pc.community_id = c.id").
		Where("p.id = ?", id).
		GroupBy("p.id", "cm.id", "person.firebase_id").
		IteratorContext(ctx).
		One(&post); err != nil {
		if err == db.ErrNoMoreRows {
			return nil, nil
		}
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
				Join("content_metadata as cm").On("p.metadata_id=cm.id").
				LeftJoin("post_communities as pc").On("p.id=pc.post_id").
				Where("(ISNULL(?) OR (cm.created_at < ? OR cm.created_at = ? AND (? = '' OR p.id < ?)))", from, from, from, cursor, cursor).
				And("(ISNULL(?) OR pc.community_id IN ?)", communityIds, communityIds).
				GroupBy("p.id")).
		As("p_ids").
		Join("post as p").On("p_ids.id = p.id").
		Join("content_metadata as cm").On("p.metadata_id = cm.id").
		Join("person").On("cm.creator_id = person.firebase_id").
		LeftJoin("post_communities as pc").On("p.id = pc.post_id").
		Join("community as c").On("pc.community_id = c.id").
		OrderBy("cm.created_at DESC", "cm.id DESC").
		GroupBy("p.id", "cm.id", "person.firebase_id").
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
		Id:              post.Id,
		ContentMetadata: buildContentMetadata(&post.flattenedMetadata),
		Title:           post.Title,
		Content:         post.Content,
		Communities:     communities,
	}, nil
}

func buildContentMetadata(metadata *flattenedMetadata) *types.ContentMetadata {
	return &types.ContentMetadata{
		Id: metadata.Id,
		Creator: &types.DisplayableUser{
			User: &types.User{
				Id:          metadata.CreatorId,
				DisplayName: metadata.CreatorDisplayName,
			},
			Alias: metadata.CreatorAlias,
		},
		NumVotes:   metadata.NumVotes,
		VoteTotal:  metadata.VoteTotal,
		Visibility: metadata.Visibility,
		CreatedAt:  metadata.CreatedAt,
		UpdatedAt:  metadata.UpdatedAt,
	}
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

func (psdb *PlanetScaleDB) Vote(ctx context.Context, userId string, targetMetadataId int64, value int8) error {
	return psdb.sess.TxContext(ctx, func(sess db.Session) error {
		row, err := sess.SQL().QueryRowContext(ctx, `SELECT value FROM vote 
																WHERE tgt_metadata_id = ? AND voter_id= ?
															FOR UPDATE`,
			targetMetadataId, userId)
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
					Where("tgt_metadata_id = ? AND voter_id = ?", targetMetadataId, userId).
					ExecContext(ctx); err != nil {
					return err
				}
				numVotesChange -= 1
			} else {
				// update existing vote
				if _, err := sess.SQL().
					Update("vote").
					Set("value", value).
					Where("tgt_metadata_id = ? AND voter_id = ?", targetMetadataId, userId).
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
				Columns("voter_id", "tgt_metadata_id", "value").
				Values(userId, targetMetadataId, value).
				ExecContext(ctx); err != nil {
				return err
			}
			numVotesChange += 1
		}

		_, err = sess.SQL().
			Update("content_metadata").
			Set("vote_total = vote_total + ?, num_votes = num_votes + ?", netVoteChange, numVotesChange).
			Where("id = ?", targetMetadataId).
			ExecContext(ctx)
		return err
	}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
}

func (psdb *PlanetScaleDB) CreateReport(ctx context.Context, userId string, req *CreateReport) (int64, error) {
	res, err := psdb.sess.SQL().
		InsertInto("report").
		Columns("tgt_metadata_id", "creator_id", "reason").
		Values(req.PostId, userId, req.Reason).
		ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
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
