package planetscale

import (
	"context"
	"database/sql"
	"encoding/json"
	appDb "github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/navbryce/next-dorm-be/util"
	"github.com/upper/db/v4"
	"time"
)

type PostDB struct {
	sess db.Session
}

func getPostDB(sess db.Session) *PostDB {
	return &PostDB{sess}
}
func (cdb *PostDB) CreatePost(ctx context.Context, post *appDb.CreatePost) (int64, error) {
	var postId int64
	err := cdb.sess.TxContext(ctx, func(sess db.Session) error {
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

// EditPost updates the post. Only updates title or content if they are non-empty.
func (cdb *PostDB) EditPost(ctx context.Context, id int64, req *appDb.EditPost) error {
	return cdb.sess.TxContext(ctx, func(sess db.Session) error {
		// TODO: Can be made more efficient if not fetching metadata id
		var metadataId struct {
			Id int64 `db:"metadata_id"`
		}
		err := sess.SQL().
			Select("metadata_id").
			From("post").
			Where("id = ?", id).
			One(&metadataId)
		if err != nil {
			return err
		}

		err = editContentMetadata(ctx, sess, metadataId.Id, req.EditContentMetadata)
		if err != nil {
			return err
		}

		if len(req.Title) > 0 || len(req.Content) > 0 {
			updater := sess.SQL().
				Update("post")
			if len(req.Title) > 0 {
				updater = updater.Set("title = ?", req.Title)
			}
			if len(req.Content) > 0 {
				updater = updater.Set("content = ?", req.Content)
			}
			if _, err := updater.ExecContext(ctx); err != nil {
				return err
			}
		}

		return err
	}, nil)
}

func (cdb *PostDB) MarkPostAsDeleted(ctx context.Context, id int64) error {
	_, err := cdb.sess.SQL().ExecContext(ctx, db.Raw(`
UPDATE post as p
	INNER JOIN content_metadata as cm ON p.metadata_id = cm.id
	SET cm.status = 'DELETED', p.content=''
	WHERE p.id = ?
`, id))
	return err
}

func (cdb *PostDB) CreateComment(ctx context.Context, req *appDb.CreateComment) (int64, error) {
	var commentId int64
	err := cdb.sess.TxContext(ctx, func(sess db.Session) error {
		metadataId, err := insertContentMetadata(ctx, sess, req.CreateContentMetadata)
		if err != nil {
			return err
		}

		res, err := sess.SQL().
			InsertInto("comment").
			Values(metadataId, req.RootMetadataId, req.ParentMetadataId, req.Content).
			Columns("metadata_id", "root_metadata_id", "parent_metadata_id", "content").
			ExecContext(ctx)
		if err != nil {
			return err
		}
		commentId, err = res.LastInsertId()

		if _, err = sess.SQL().
			Update("post").
			Set("comment_count = comment_count + 1").
			Where("metadata_id = ?", req.ParentMetadataId).
			ExecContext(ctx); err != nil {
			return err
		}
		return err
	}, &sql.TxOptions{})
	return commentId, err
}

func (cdb *PostDB) MarkCommentAsDeleted(ctx context.Context, id int64) error {
	_, err := cdb.sess.SQL().ExecContext(ctx, db.Raw(`
UPDATE comment as c
	INNER JOIN content_metadata as cm ON c.metadata_id = cm.id
	WHERE c.id = ?
	SET cm.status = 'DELETED', c.content=''
`, id))
	return err
}

func insertContentMetadata(ctx context.Context, sess db.Session, metadata *appDb.CreateContentMetadata) (id int64, err error) {
	if err != nil {
		return 0, err
	}

	res, err := sess.SQL().
		InsertInto("content_metadata").
		Columns("creator_id", "creator_alias", "visibility").
		Values(metadata.CreatorId, metadata.CreatorAlias, metadata.Visibility).
		ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, insertContentImagesFor(ctx, sess, id, metadata.ImageBlobNames)
}

func editContentMetadata(ctx context.Context, sess db.Session, metadataId int64, req *appDb.EditContentMetadata) error {
	if _, err := sess.SQL().ExecContext(ctx, db.Raw(`
		DELETE ci FROM content_image ci
		JOIN image i ON ci.image_id = i.id
		WHERE ci.metadata_id = ? AND JSON_CONTAINS(?, JSON_QUOTE(i.blob_name))
	`, metadataId, req.ImageBlobNamesToRemove)); err != nil {
		return err
	}

	err := insertContentImagesFor(ctx, sess, metadataId, req.ImageBlobNamesToAdd)
	if err != nil {
		return nil
	}

	_, err = sess.SQL().
		Update("content_metadata").
		Set("visibility = ?", req.Visibility).
		ExecContext(ctx)
	return err
}

// TODO: Move to separate file? Simplify image model if not extended upon
func insertContentImagesFor(ctx context.Context, sess db.Session, metadataId int64, imageBlobNames []string) error {
	imageIds := make([]int64, len(imageBlobNames))
	for i, imageBlobName := range imageBlobNames {
		res, err := sess.SQL().
			InsertInto("image").
			Columns("blob_name").
			Values(imageBlobName).
			ExecContext(ctx)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		imageIds[i] = id
	}

	batchInserter := sess.SQL().
		InsertInto("content_image").
		Columns("metadata_id", "image_id").
		Batch(len(imageIds))
	for _, imageId := range imageIds {
		batchInserter.Values(metadataId, imageId)
	}
	batchInserter.Done()
	if err := batchInserter.Wait(); err != nil {
		return err
	}
	return nil
}

type flattenedUserVote struct {
	Value sql.NullInt16 `db:"value"`
}

type flattenedContentMetadata struct {
	Id                 int64            `db:"metadata_id"`
	NumVotes           int              `db:"num_votes"`
	VoteTotal          int              `db:"vote_total"`
	CreatorId          string           `db:"creator_id"`
	CreatorDisplayName string           `db:"display_name"`
	CreatorAlias       string           `db:"creator_alias"`
	Visibility         model.Visibility `db:"visibility"`
	Status             model.Status     `db:"status"`
	flattenedUserVote  `db:",inline"`
	ImageBlobNamesStr  string    `db:"image_blob_names"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

type flattenedPost struct {
	flattenedContentMetadata `db:",inline"`
	Id                       int64  `db:"id"`
	Title                    string `db:"title"`
	Content                  string `db:"content"`
	CommunityIdsStr          string `db:"community_ids"`
	CommunityNamesJSONStr    string `db:"community_names"`
	CommentCount             int64  `db:"comment_count"`
}

var contentMetadataColumns = []interface{}{
	"cm.id as metadata_id",
	"cm.creator_id",
	"person.display_name",
	"cm.creator_alias",
	"cm.num_votes",
	"cm.vote_total",
	"cm.visibility",
	"cm.status",
	"cm.created_at",
	"cm.updated_at",
}

var postColumns = append(contentMetadataColumns,
	[]interface{}{
		"p.id",
		"p.title",
		"p.content",
		"p.comment_count",
		db.Raw("JSON_ARRAYAGG(image.blob_name) as image_blob_names"),
		db.Raw("JSON_ARRAYAGG(pc.community_id) as community_ids"), db.Raw("JSON_ARRAYAGG(c.name) as community_names"),
	}...)

var voteColumns = []interface{}{
	"v.value",
}

func (cdb *PostDB) GetPostById(ctx context.Context, id int64, opts *appDb.PostQueryOpts) (*model.Post, error) {
	var post flattenedPost
	if err := cdb.sess.SQL().
		Select(append(postColumns, voteColumns...)...).
		From("post AS p").
		Join("content_metadata as cm").On("p.metadata_id = cm.id").
		LeftJoin("post_communities as pc").On("p.id = pc.post_id").
		// TODO: This can be optimized: don't join if VoteHistoryOf empty
		LeftJoin("vote as v").On("v.voter_id = ? AND cm.id = v.tgt_metadata_id", opts.VoteHistoryOf).
		Join("person").On("cm.creator_id = person.firebase_id").
		Join("community as c").On("pc.community_id = c.id").
		LeftJoin("content_image as ci").On("cm.id = ci.metadata_id").
		LeftJoin("image").On("ci.image_id = image.id").
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

func (cdb *PostDB) GetPosts(ctx context.Context, query *appDb.PostsListQuery) ([]*model.Post, error) {
	if query.CommunityIds != nil && len(query.CommunityIds) == 0 {
		return []*model.Post{}, nil
	}
	var optionalConditions []*db.RawExpr

	// TODO: Convert to an interface
	if query.From != nil {
		optionalConditions = append(optionalConditions,
			db.Raw("(cm.created_at < ? OR cm.created_at = ? AND (? = '' OR p.id < ?))",
				query.From, query.From, query.LastId, query.LastId),
		)
	}
	if query.CommunityIds != nil {
		optionalConditions = append(optionalConditions, db.Raw("(pc.community_id IN ?)", query.CommunityIds))
	}

	if query.ByUser != nil {
		optionalConditions = append(optionalConditions, db.Raw("(cm.creator_id = ?)", query.ByUser.Id))
	}

	if !query.IncludeDeleted {
		optionalConditions = append(optionalConditions, db.Raw("(cm.status != 'DELETED')"))
	}

	var flattenedPosts []flattenedPost
	if err := cdb.sess.SQL().
		Select(append(postColumns, voteColumns...)...).
		From(
			cdb.sess.SQL().
				Select("p.id").
				From("post as p").
				Join("content_metadata as cm").On("p.metadata_id=cm.id").
				LeftJoin("post_communities as pc").On("p.id=pc.post_id").
				Where(andExpressions(optionalConditions...)...).
				And("(? OR pc.community_id IN ?)", query.CommunityIds == nil, query.CommunityIds).
				GroupBy("p.id")).
		As("p_ids").
		Join("post as p").On("p_ids.id = p.id").
		Join("content_metadata as cm").On("p.metadata_id = cm.id").
		// TODO: This can be optimized: don't join if VoteHistoryOf empty
		LeftJoin("vote as v").On("v.voter_id = ? AND cm.id = v.tgt_metadata_id", query.VoteHistoryOf).
		Join("person").On("cm.creator_id = person.firebase_id").
		LeftJoin("post_communities as pc").On("p.id = pc.post_id").
		Join("community as c").On("pc.community_id = c.id").
		LeftJoin("content_image as ci").On("cm.id = ci.metadata_id").
		LeftJoin("image").On("ci.image_id = image.id").
		OrderBy("cm.created_at DESC", "cm.id DESC").
		GroupBy("p.id", "cm.id", "person.firebase_id").
		Limit(int(query.Limit)).
		IteratorContext(ctx).
		All(&flattenedPosts); err != nil {
		return nil, err
	}
	posts := make([]*model.Post, len(flattenedPosts))
	for i, flattened := range flattenedPosts {
		post, err := buildPostFromFlattened(&flattened)
		if err != nil {
			return nil, err
		}
		posts[i] = post
	}
	return posts, nil
}

func buildPostFromFlattened(post *flattenedPost) (*model.Post, error) {
	var communityIds []int64
	if err := json.Unmarshal([]byte(post.CommunityIdsStr), &communityIds); err != nil {
		return nil, err
	}

	var communityNames []string
	if err := json.Unmarshal([]byte(post.CommunityNamesJSONStr), &communityNames); err != nil {
		return nil, err
	}

	communities := make([]*model.Community, len(communityIds))
	for i, communityId := range communityIds {
		communities[i] = &model.Community{
			Id:   communityId,
			Name: communityNames[i],
		}
	}
	metadata, err := buildContentMetadataFromFlattened(&post.flattenedContentMetadata)
	if err != nil {
		return nil, err
	}

	return &model.Post{
		Id:              post.Id,
		ContentMetadata: metadata,
		Title:           post.Title,
		Content:         post.Content,
		Communities:     communities,
		CommentCount:    post.CommentCount,
	}, nil
}

type flattenedComment struct {
	flattenedContentMetadata `db:",inline"`
	Id                       int64  `db:"id"`
	RootMetadataId           int64  `db:"root_metadata_id"`
	ParentMetadataId         int64  `db:"parent_metadata_id"`
	Content                  string `db:"content"`
}

var commentColumns = append([]interface{}{
	"c.id",
	"c.metadata_id",
	"c.root_metadata_id",
	"c.parent_metadata_id",
	"c.content",
}, contentMetadataColumns...)

func (cdb *PostDB) GetCommentById(ctx context.Context, id int64) (*model.Comment, error) {
	var comment flattenedComment
	if err := cdb.sess.SQL().
		Select(commentColumns...).
		From("comment as c").
		Join("content_metadata as cm").On("c.metadata_id = cm.id").
		Join("person").On("cm.creator_id = person.firebase_id").
		Where("c.id = ?", id).
		IteratorContext(ctx).
		One(&comment); err != nil {
		if err == db.ErrNoMoreRows {
			return nil, nil
		}
		return nil, err
	}
	return buildCommentFromFlattened(&comment)
}

func (cdb *PostDB) GetCommentForest(ctx context.Context, rootMetadataId int64, opts *appDb.CommentTreeQueryOpts) ([]*model.CommentTree, error) {
	var flattenedComments []flattenedComment
	if err := cdb.sess.SQL().
		Select(append(commentColumns, voteColumns...)...).
		From("comment as c").
		Join("content_metadata as cm").On("c.metadata_id = cm.id").
		// TODO: This can be optimized: don't join if VoteHistoryOf empty
		LeftJoin("vote as v").On("v.voter_id = ? AND cm.id = v.tgt_metadata_id", opts.VoteHistoryOf).
		Join("person").On("cm.creator_id = person.firebase_id").
		Where("root_metadata_id = ?", rootMetadataId).
		OrderBy("created_at").
		IteratorContext(ctx).
		All(&flattenedComments); err != nil {
		return nil, err
	}

	comments := make([]*model.Comment, len(flattenedComments))
	for i, flattenedComment := range flattenedComments {
		comment, err := buildCommentFromFlattened(&flattenedComment)
		if err != nil {
			return nil, err
		}
		comments[i] = comment
	}

	return buildCommentForest(rootMetadataId, comments), nil
}

func buildCommentFromFlattened(comment *flattenedComment) (*model.Comment, error) {
	metadata, err := buildContentMetadataFromFlattened(&comment.flattenedContentMetadata)
	if err != nil {
		return nil, err
	}
	return &model.Comment{
		Id:               comment.Id,
		ContentMetadata:  metadata,
		RootMetadataId:   comment.RootMetadataId,
		ParentMetadataId: comment.ParentMetadataId,
		Content:          comment.Content,
	}, nil
}

func buildContentMetadataFromFlattened(metadata *flattenedContentMetadata) (*model.ContentMetadata, error) {
	var vote *model.Vote
	if metadata.flattenedUserVote.Value.Valid {
		value, err := metadata.flattenedUserVote.Value.Value()
		if err != nil {
			panic(err)
		}
		vote = &model.Vote{Value: int8(value.(int64))}
	}
	var imageBlobNames = []string{} // make sure the array isn't nil
	if len(metadata.ImageBlobNamesStr) > 0 {
		if err := json.Unmarshal([]byte(metadata.ImageBlobNamesStr), &imageBlobNames); err != nil {
			return nil, err
		}
		// remove empty blob names (nulls)
		for i, blobName := range imageBlobNames {
			if len(blobName) == 0 {
				imageBlobNames = append(imageBlobNames[0:i], imageBlobNames[i+1:]...)
			}
		}
	}

	return &model.ContentMetadata{
		Id: metadata.Id,
		Creator: &model.DisplayableUser{
			User: &model.User{
				Id:          metadata.CreatorId,
				DisplayName: metadata.CreatorDisplayName,
				Avatar:      util.Avatar(metadata.CreatorId),
			},
			AnonymousUser: &model.AnonymousUser{
				DisplayName: metadata.CreatorAlias,
				Avatar:      util.Avatar(metadata.CreatorAlias),
			},
		},
		UserVote:       vote,
		Status:         metadata.Status,
		NumVotes:       metadata.NumVotes,
		VoteTotal:      metadata.VoteTotal,
		Visibility:     metadata.Visibility,
		ImageBlobNames: imageBlobNames,
		CreatedAt:      metadata.CreatedAt,
		UpdatedAt:      metadata.UpdatedAt,
	}, nil
}

func buildCommentForest(rootId int64, comments []*model.Comment) []*model.CommentTree {
	adj := make(map[int64][]*model.Comment)
	for _, comment := range comments {
		adj[comment.ParentMetadataId] = append(adj[comment.ParentMetadataId], comment)
	}
	return buildCommentForestFromAdjList(adj, rootId)
}

func buildCommentForestFromAdjList(adj map[int64][]*model.Comment, rootId int64) []*model.CommentTree {
	comments, ok := adj[rootId]
	if !ok {
		return []*model.CommentTree{}
	}
	forest := make([]*model.CommentTree, len(comments))
	for i, comment := range comments {
		forest[i] = &model.CommentTree{
			Comment:  comment,
			Children: buildCommentForestFromAdjList(adj, comment.ContentMetadata.Id),
		}
	}
	return forest
}
func (cdb *PostDB) Vote(ctx context.Context, userId string, targetMetadataId int64, value int8) error {
	return cdb.sess.TxContext(ctx, func(sess db.Session) error {
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

func (cdb *PostDB) CreateReport(ctx context.Context, userId string, req *appDb.CreateReport) (int64, error) {
	res, err := cdb.sess.SQL().
		InsertInto("report").
		Columns("tgt_metadata_id", "creator_id", "reason").
		Values(req.PostId, userId, req.Reason).
		ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
