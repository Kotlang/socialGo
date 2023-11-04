package extensions

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/Kotlang/socialGo/models"
	"github.com/jinzhu/copier"
	"golang.org/x/net/html"
)

// func GetSubscribedPostIds(db *db.SocialDb, tenant string, subscriberId string) chan []string {
// 	postIds := make(chan []string)

// 	go func() {
// 		likeFilters := bson.M{
// 			"userId":   subscriberId,
// 			"postType": pb.PostType_SOCIAL_EVENT.String(),
// 		}

// 		likeCountChan, errChan := db.PostLike(tenant).CountDocuments(likeFilters)
// 		var count int64 = 0
// 		select {
// 		case likeCount := <-likeCountChan:
// 			count = likeCount
// 		case err := <-errChan:
// 			logger.Error("Failed getting subscribed post count", zap.Error(err))
// 			postIds <- []string{}
// 			return
// 		}

// 		likePostsChan, errChan := db.PostLike(tenant).Find(likeFilters, bson.D{}, count, 0)
// 		likePostIds := []string{}
// 		select {
// 		case likePosts := <-likePostsChan:
// 			for _, likePost := range likePosts {
// 				likePostIds = append(likePostIds, likePost.PostId)
// 			}
// 			postIds <- likePostIds
// 		case err := <-errChan:
// 			logger.Error("Failed getting like posts", zap.Error(err))
// 			postIds <- []string{}
// 			return
// 		}
// 	}()

// 	return postIds
// }

func SaveTags(db *db.SocialDb, tenant string, tags []string) chan bool {
	savedTagsPromise := make(chan bool)

	go func() {
		for _, tag := range tags {
			existingTagChan, errChan := db.Tag(tenant).FindOneById(tag)
			select {
			case existingTag := <-existingTagChan:
				existingTag.NumPosts++
				<-db.Tag(tenant).Save(existingTag)
			case <-errChan:
				newTag := &models.PostTagModel{
					Tag:      tag,
					NumPosts: 1,
				}
				<-db.Tag(tenant).Save(newTag)
			}
		}

		savedTagsPromise <- true
	}()

	return savedTagsPromise
}

// Adds additional userProfile data, comments/answers to feedPost parameter.
func AttachPostUserInfoAsync(
	socialDb *db.SocialDb,
	grpcContext context.Context,
	feedPost *pb.UserPostProto,
	userId, tenant, userType string,
	attachAnswers bool) chan bool {

	// logger.Info("AttachPostUserInfoAsync", zap.Any("feedPost", feedPost))

	done := make(chan bool)

	go func() {
		feedPost.HasFeedUserLiked = socialDb.PostLike(tenant).IsExistsById(
			(&models.PostLikeModel{UserId: userId, PostId: feedPost.PostId}).Id(),
		)

		// get post author profile
		authorProfile := <-GetSocialProfile(grpcContext, feedPost.UserId)
		feedPost.AuthorInfo = authorProfile

		if attachAnswers {
			answers := socialDb.FeedPost(tenant).GetFeed(
				nil, // since we filter by parent PostId, we don't need to filter by postType.
				feedPost.PostId,
				int64(0),
				int64(10))
			answersProto := []*pb.UserPostProto{}
			copier.Copy(&answersProto, answers)
			feedPost.AnswersThread = answersProto

			// recursively attach authorInfo to answers.
			for _, answerProto := range feedPost.AnswersThread {
				<-AttachPostUserInfoAsync(socialDb, grpcContext, answerProto, userId, tenant, userType, false)
			}
		}

		done <- true
	}()

	return done
}

type Regex struct {
	Youtube *regexp.Regexp
	Links   *regexp.Regexp
}

var rg *Regex

const linksExpr string = `(https?:\/\/[^\s]+)`
const youtubeExpr string = `^.*(?:(?:youtu\.be\/|v\/|vi\/|u\/\w\/|embed\/|shorts\/)|(?:(?:watch)?\?v(?:i)?=|\&v(?:i)?=))([^#\&\?]*).*`

func GetLinks(content string) chan []string {
	linksChan := make(chan []string)
	if rg == nil {
		rg = &Regex{}
		rg.Links, _ = regexp.Compile(linksExpr)
		rg.Youtube, _ = regexp.Compile(youtubeExpr)
	}
	go func() {
		linksChan <- rg.Links.FindAllString(content, -1)
	}()
	return linksChan
}

func GeneratePreviews(urls []string) (chan []*pb.MediaUrl, chan []*pb.WebPreview) {
	mediaUrlsChan := make(chan []*pb.MediaUrl)
	webPreviewsChan := make(chan []*pb.WebPreview)
	go func() {
		mediaUrls := []*pb.MediaUrl{}
		webPreviews := []*pb.WebPreview{}
		wg := &sync.WaitGroup{}
		mut := &sync.RWMutex{}
		for _, url := range urls {
			if subMatch := rg.Youtube.FindStringSubmatch(url); len(subMatch) > 1 {
				mediaUrls = append(mediaUrls, &pb.MediaUrl{Url: subMatch[1], MimeType: "video/x-youtube"})
			}
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				mut.Lock()
				webPreviews = append(webPreviews, generateWebPreview(url))
				mut.Unlock()
			}(url)
		}
		wg.Wait()
		mediaUrlsChan <- mediaUrls
		mut.RLock()
		webPreviewsChan <- webPreviews
		mut.RUnlock()
	}()
	return mediaUrlsChan, webPreviewsChan
}

func generateWebPreview(url string) *pb.WebPreview {
	webPreview := &pb.WebPreview{Url: url}

	res, err := http.Get(url)
	if err != nil {
		return webPreview
	}
	defer res.Body.Close()

	doc, err := html.Parse(res.Body)
	if err != nil {
		return webPreview
	}
	traverse(doc, webPreview)
	return webPreview
}

func traverse(n *html.Node, webPreview *pb.WebPreview) {
	if n == nil || (n.Type == html.ElementNode && n.Data == "body") {
		return
	}
	if n.Type == html.ElementNode && n.Data == "title" && len(webPreview.Title) == 0 {
		webPreview.Title = strings.TrimSpace(n.FirstChild.Data)
	}
	if n.Type == html.ElementNode && n.Data == "meta" {
		for _, attr := range n.Attr {
			switch attr.Key {
			case "name":
				if attr.Val == "description" && len(webPreview.Description) == 0 {
					content := getContent(n.Attr)
					if len(content) > 0 {
						webPreview.Description = content
					}
				}
			case "property":
				if attr.Val == "og:title" && len(webPreview.Title) == 0 {
					content := getContent(n.Attr)
					if len(content) > 0 {
						webPreview.Title = content
					}
				} else if attr.Val == "og:image" {
					content := getContent(n.Attr)
					if len(content) > 0 {
						webPreview.PreviewImage = content
					}
				} else if attr.Val == "og:description" && len(webPreview.Description) == 0 {
					content := getContent(n.Attr)
					if len(strings.TrimSpace(content)) > 0 {
						webPreview.Description = content
					}
				}

			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverse(c, webPreview)
	}
}

func getContent(attributes []html.Attribute) string {
	for _, attr := range attributes {
		if attr.Key == "content" {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}
