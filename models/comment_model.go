package models

type CommentModel struct {
	CommentId   string           `bson:"_id"`
	Content     string           `bson:"content"`
	MediaUrls   []MediaUrl       `bson:"mediaUrls"`
	WebPreviews []WebPreview     `bson:"webPreviews"`
	ParentId    string           `bson:"parentId"`
	UserId      string           `bson:"userId"`
	CommentOn   string           `bson:"commentOn"`
	NumReacts   map[string]int64 `bson:"numReacts"`
	NumReplies  int              `bson:"numReplies"`
	NumShares   int              `bson:"numShares"`
	CreatedOn   int64            `bson:"createdOn"`
	IsDeleted   bool             `bson:"isDeleted"`
}

func (c *CommentModel) Id() string {
	if len(c.CommentId) == 0 {
		c.CommentId = GetCommentId(c.UserId, c.ParentId)
	}
	return c.CommentId
}

// returns the comment id for the given user and parent
func GetCommentId(userId, parentId string) string {
	return userId + "/" + parentId
}
