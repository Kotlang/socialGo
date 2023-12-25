package extensions

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/Kotlang/socialGo/db"
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/Kotlang/socialGo/models"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/net/html"
)

func SaveTags(db db.SocialDbInterface, tenant string, tags []string) chan bool {
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

// Adds additional userProfile data, reactions data to feedPost proto.
func AttachPostUserInfoAsync(
	socialDb db.SocialDbInterface,
	grpcContext context.Context,
	feedPost *socialPb.UserPostProto,
	userId, tenant string) chan bool {

	done := make(chan bool)

	go func() {
		// get user reactions for the post
		feedPost.FeedUserReactions = socialDb.React(tenant).GetUserReactions(userId, feedPost.PostId)
		// get post author profile
		authorProfile := <-GetSocialProfile(grpcContext, feedPost.UserId)
		feedPost.AuthorInfo = authorProfile

		done <- true
	}()

	return done
}

// Adds additional userProfile data, reactions data to multiple feedPost proto.
func AttachMultiplePostUserInfoAsync(
	socialDb db.SocialDbInterface,
	grpcContext context.Context,
	feedPosts []*socialPb.UserPostProto,
	userId, tenant string) chan bool {

	done := make(chan bool)

	authorIds := []string{}
	entityIds := []string{}
	for _, feedPost := range feedPosts {
		authorIds = append(authorIds, feedPost.UserId)
		// list of reaction ids for the post
		entityIds = append(entityIds, models.GetReactionId(userId, feedPost.PostId))
	}
	go func() {
		filter := bson.M{
			"_id": bson.M{
				"$in": entityIds,
			},
		}
		reactionResChan, errChan := socialDb.React(tenant).Find(filter, bson.D{}, 0, 0)
		authorProfilesPromise := GetSocialProfiles(grpcContext, authorIds)

		select {
		case reactions := <-reactionResChan:
			reactionsMap := make(map[string][]string)
			for _, reaction := range reactions {
				reactionsMap[reaction.EntityId] = reaction.Reaction
			}
			for _, feedPost := range feedPosts {
				feedPost.FeedUserReactions = reactionsMap[feedPost.PostId]
				if feedPost.FeedUserReactions == nil {
					feedPost.FeedUserReactions = []string{}
				}
			}
		case <-errChan:
		}

		authorProfiles := <-authorProfilesPromise
		authorProfilesMap := make(map[string]*socialPb.SocialProfile)
		for _, authorProfile := range authorProfiles {
			authorProfilesMap[authorProfile.UserId] = authorProfile
		}
		for _, feedPost := range feedPosts {
			feedPost.AuthorInfo = authorProfilesMap[feedPost.UserId]
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

func GeneratePreviews(urls []string) (chan []*socialPb.MediaUrl, chan []*socialPb.WebPreview) {
	mediaUrlsChan := make(chan []*socialPb.MediaUrl)
	webPreviewsChan := make(chan []*socialPb.WebPreview)
	go func() {
		mediaUrls := []*socialPb.MediaUrl{}
		webPreviews := []*socialPb.WebPreview{}
		wg := &sync.WaitGroup{}
		mut := &sync.RWMutex{}
		for _, url := range urls {
			if subMatch := rg.Youtube.FindStringSubmatch(url); len(subMatch) > 1 {
				mediaUrls = append(mediaUrls, &socialPb.MediaUrl{Url: subMatch[1], MimeType: "video/x-youtube"})
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

func generateWebPreview(url string) *socialPb.WebPreview {
	webPreview := &socialPb.WebPreview{Url: url}

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

func traverse(n *html.Node, webPreview *socialPb.WebPreview) {
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
