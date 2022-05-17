package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/navbryce/next-dorm-be/db"
	"github.com/navbryce/next-dorm-be/model"
	"github.com/navbryce/next-dorm-be/util"
	"log"
	"sync"
	"time"
)

type communityTree struct {
	adjList             map[int64][]*model.Community
	parentAdjList       map[int64]*model.Community
	mostRecentCommunity *time.Time
	createdAt           *time.Time
}

func (ct *communityTree) isNewer(tree *communityTree) bool {
	return ct.mostRecentCommunity.After(*tree.mostRecentCommunity)
}

const TreeUpdateInterval = time.Minute * 20

type CommunityController struct {
	db             db.CommunityDatabase
	cachedTree     *communityTree
	cachedTreeLock sync.Mutex
	updateTicker   *time.Ticker
}

func NewCommunityController(c context.Context, db db.CommunityDatabase) (*CommunityController, error) {
	controller := &CommunityController{
		db: db,
	}
	if err := controller.updateCachedTree(c); err != nil {
		return nil, err
	}

	// TODO: Be careful of loops syncing?
	updateTicker := time.NewTicker(TreeUpdateInterval)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("recovered while attempting to update cached tree", r)
			}
		}()
		for {
			select {
			case <-updateTicker.C:
				controller.attemptToUpdateCachedTree(c)
			}
		}
	}()

	return controller, nil
}

func (cc *CommunityController) CreateCommunity(c context.Context, name string) (int64, *util.HTTPError) {
	community, err := cc.db.CreateCommunity(c, name)
	if err != nil {
		return -1, util.BuildDbHTTPErr(err)
	}
	go cc.attemptToUpdateCachedTree(c)

	return community, nil
}

func (cc *CommunityController) GetCommunityById(c context.Context, id int64, opts *db.GetCommunitiesQueryOpts) (*model.CommunityWithSubStatus, *util.HTTPError) {
	communities, err := cc.db.GetCommunitiesByIds(c, []int64{id}, opts)
	if err != nil {
		return nil, util.BuildDbHTTPErr(err)
	}
	return communities[0], nil
}

func (cc *CommunityController) GetCommunityPos(c *gin.Context, id int64) (*model.CommunityPosInTree, *util.HTTPError) {
	children := []*model.Community{}
	if cc.cachedTree.adjList[id] != nil {
		children = cc.cachedTree.adjList[id]
	}

	parents := []*model.Community{} // DON'T return nil slice
	for parent := cc.cachedTree.parentAdjList[id]; parent != nil; parent = cc.cachedTree.parentAdjList[parent.Id] {
		parents = append(parents, parent)
	}

	return &model.CommunityPosInTree{
		Children: children,
		Path:     parents,
	}, nil
}

func (cc *CommunityController) attemptToUpdateCachedTree(c context.Context) {
	if err := cc.updateCachedTree(c); err != nil {
		log.Println("an error occurred while updating the cached tree", err)
	}
}

func (cc *CommunityController) updateCachedTree(c context.Context) error {
	allCommunities, err := cc.db.GetCommunitiesByIds(c, nil, &db.GetCommunitiesQueryOpts{})
	if err != nil {
		return err
	}
	newTree := buildTreeFromCommunities(communitiesWithSubStatusesToCommunity(allCommunities))

	// start of cachedTreeLock
	cc.cachedTreeLock.Lock()
	defer cc.cachedTreeLock.Unlock()
	if cc.cachedTree == nil || newTree.isNewer(cc.cachedTree) {
		cc.cachedTree = newTree
	}
	// end of cachedTreeLock
	return nil
}

func communitiesWithSubStatusesToCommunity(communitiesWithStatuses []*model.CommunityWithSubStatus) []*model.Community {
	communities := make([]*model.Community, len(communitiesWithStatuses))
	for i, community := range communitiesWithStatuses {
		communities[i] = community.Community
	}
	return communities
}

func buildTreeFromCommunities(communities []*model.Community) *communityTree {
	var mostRecent *time.Time
	adjList := make(map[int64][]*model.Community)
	idToCommunity := make(map[int64]*model.Community)
	parentAdjList := make(map[int64]*model.Community)
	for _, community := range communities {
		idToCommunity[community.Id] = community
		if mostRecent == nil || community.CreatedAt.After(*mostRecent) {
			mostRecent = community.CreatedAt
		}
		adjList[community.ParentId.AsInt()] = append(adjList[community.ParentId.AsInt()], community)
	}

	for _, community := range communities {
		parentAdjList[community.Id] = idToCommunity[community.ParentId.AsInt()]
	}
	now := time.Now()
	return &communityTree{
		createdAt:           &now,
		adjList:             adjList,
		parentAdjList:       parentAdjList,
		mostRecentCommunity: mostRecent,
	}
}
