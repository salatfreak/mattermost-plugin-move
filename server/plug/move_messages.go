package plug

import (
  "time"
  "encoding/json"

  "github.com/mattermost/mattermost-server/v6/model"

  "github.com/salatfreak/mattermost-plugin-move/server/i18n"
)

func (p *Plug) runMoveMessages(
  teamId, channelId, targetPostId, userId string, sourcePostIds []string,
) error {
  p.api.Log.Debug(
    "Running move command",
    "team", teamId, "channel", channelId, "targetPost", targetPostId,
    "user", userId, "sourcePosts", sourcePostIds,
  )

  // Get target channel and post
  tgtChannel, err := p.api.Channel.Get(channelId)
  if err != nil { return err }
  var tgtPost *model.Post
  if targetPostId != "" {
    tgtPost, err = p.api.Post.GetPost(targetPostId)
    if err != nil { return err }
  }

  // Get source posts
  srcPosts, err := p.getPostsFromIds(sourcePostIds)
  if err != nil { return err }

  // Check permissions
  for _, post := range(srcPosts) {
    if tgtPost != nil && post.Id == tgtPost.Id {
      return i18n.NewError(i18n.MsgErrorAttachItself)
    }
    if tgtPost != nil && post.CreateAt <= tgtPost.CreateAt {
      return i18n.NewError(i18n.MsgErrorNewerMessage)
    }
    err := p.assertSourcePermissions(userId, post, tgtChannel)
    if err != nil { return err }
  }
  err = p.assertTargetPermissions(userId, tgtChannel)
  if err != nil { return err }

  // Move messages
  for _, post := range(srcPosts) {
    err = p.movePost(userId, post, tgtChannel, tgtPost)
    if err != nil { return err }
  }
  return nil
}

func (p *Plug) getPostsFromIds(postIds []string) ([]*model.Post, error) {
  posts := make([]*model.Post, 0, len(postIds))
  for _, postId := range(postIds) {
    post, err := p.api.Post.GetPost(postId)
    if err != nil {
      return nil, i18n.NewError(i18n.MsgErrorNotExist, "PostId", postId)
    }
    posts = append(posts, post)
  }
  return posts, nil
}

func (p *Plug) assertSourcePermissions(
  userId string, post *model.Post, tgtChannel *model.Channel,
) error {
  p.api.Log.Debug("Checking source permissions")

  // Check message delete permission
  var perm *model.Permission
  switch post.UserId {
    case userId: perm = model.PermissionDeletePost
    default: perm = model.PermissionDeleteOthersPosts
  }
  if !p.api.User.HasPermissionToChannel(userId, post.ChannelId, perm) {
    return i18n.NewError(i18n.MsgErrorPermissionMessage, "PostId", post.Id)
  }

  // Check thread delete permission
  if post.RootId == "" {
    threadPosts, err := p.getThreadPosts(post.Id)
    if err != nil { return err }
    for _, reply := range(threadPosts) {
      var perm *model.Permission
      switch reply.UserId {
        case userId: perm = model.PermissionDeletePost
        default: perm = model.PermissionDeleteOthersPosts
      }
      if !p.api.User.HasPermissionToChannel(userId, reply.ChannelId, perm) {
        return i18n.NewError(i18n.MsgErrorPermissionReplies, "PostId", post.Id)
      }
    }
  }

  // Check channel restrictions
  if post.ChannelId != tgtChannel.Id {
    srcChannel, err := p.api.Channel.Get(post.ChannelId)
    if err != nil { return err }
    if srcChannel.TeamId != tgtChannel.TeamId {
      return i18n.NewError(i18n.MsgErrorOtherTeam)
    }
    if srcChannel.Type != model.ChannelTypeOpen {
      return i18n.NewError(i18n.MsgErrorPrivateChannel)
    }
  }
  return nil
}

func (p *Plug) getThreadPosts(
  postId string,
) ([]*model.Post, error) {
  thread, err := p.api.Post.GetPostThread(postId)
  if err != nil { return nil, err }
  thread.UniqueOrder()
  thread.SortByCreateAt() // This strangely sorts descendingly
  for i, j := 0, len(thread.Order) - 1; i < j; i, j = i + 1, j - 1 {
    thread.Order[i], thread.Order[j] = thread.Order[j], thread.Order[i]
  }
  return thread.ToSlice(), nil
}

func (p *Plug) assertTargetPermissions(
  userId string, tgtChannel *model.Channel,
) error {
  p.api.Log.Debug("Checking target permissions")
  if !p.api.User.HasPermissionToChannel(
    userId, tgtChannel.Id, model.PermissionCreatePost,
  ) {
    name := tgtChannel.DisplayName
    return i18n.NewError(i18n.MsgErrorPermissionTarget, "ChannelName", name)
  }
  return nil
}

func (p *Plug) movePost(
  userId string, source *model.Post, channel *model.Channel, root *model.Post,
) error {
  p.api.Log.Debug(
    "Moving post", "post", source, "channel", channel, "thread", root,
  )

  // Construct post list
  var posts []*model.Post
  if source.RootId == "" {
    var err error
    posts, err = p.getThreadPosts(source.Id)
    if err != nil { return err }
  } else {
    posts = []*model.Post{ source }
  }
  p.api.Log.Debug("Retrieved post list", "posts", posts)

  // Copy posts
  for _, post := range posts {
    // Copy post
    newPost := post.Clone()
    newPost.ChannelId, newPost.RootId, newPost.Id = channel.Id, "", ""
    newPost.ReplyCount = 0
    if root != nil { newPost.RootId = root.Id }

    // Copy attachments
    newPost.FileIds = make([]string, 0, len(post.FileIds))
    for _, fileId := range post.FileIds {
      // Load old file
      info, err := p.api.File.GetInfo(fileId)
      if err != nil { return err }
      bytes, err := p.api.File.Get(fileId)
      if err != nil { return err }

      // Create new file
      newInfo, err := p.api.File.Upload(bytes, channel.Id, info.Name)
      if err != nil { return err }
      newPost.FileIds = append(newPost.FileIds, newInfo.Id)
    }
    p.api.Log.Debug("Copied attachments", "attachments", newPost.FileIds)

    // Add move event to history
    addMoveHistoryElement(newPost, map[string]any{
      "timestamp": time.Now().Unix(),
      "by_user": userId,
      "from_channel": post.ChannelId,
      "from_thread": post.RootId,
    })

    // Create new post
    err := p.api.Post.CreatePost(newPost)
    if err != nil { return err }
    p.api.Log.Debug("Created new post", "post", newPost)

    // Copy reactions
    reactions, err := p.api.Post.GetReactions(post.Id)
    if err != nil { return err }
    for _, reaction := range(reactions) {
      reaction.PostId = newPost.Id
      err := p.api.Post.AddReaction(reaction)
      if err != nil { return err }
      p.api.Log.Debug("Added reaction", "reaction", reaction)
    }

    // Set root post if nil
    if root == nil {
      root = newPost
      p.api.Log.Debug("Set post as root for further posts")
    }
  }
  
  // Delete original post
  err := p.api.Post.DeletePost(source.Id)
  if err != nil { return err }
  p.api.Log.Debug("Deleted original post")
  return nil
}

func addMoveHistoryElement(post *model.Post, element map[string]any) error {
  // Get and deserialize history
  var history []map[string]any
  if historyString, ok := post.GetProp("move_history").(string); ok {
    err := json.Unmarshal([]byte(historyString), &history)
    if err != nil { return err }
  }
  if history == nil { history = make([]map[string]any, 0, 1) }

  // Append element
  history = append(history, element)

  // Serialize and set history
  historyBytes, err := json.Marshal(history)
  if err != nil { return err }
  post.AddProp("move_history", string(historyBytes))

  return nil
}
