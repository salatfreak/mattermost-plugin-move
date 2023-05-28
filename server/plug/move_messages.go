package plug

import (
  "github.com/mattermost/mattermost-server/v6/model"

  "github.com/salatfreak/mattermost-plugin-move/server/i18n"
)

func (p *Plug) runMoveMessages(
  teamId, channelId, targetPostId, userId string, sourcePostIds []string,
) error {
  p.API.LogDebug(
    "Running move command",
    "team", teamId, "channel", channelId, "targetPost", targetPostId,
    "user", userId, "sourcePosts", sourcePostIds,
  )

  // Get target channel and post
  tgtChannel, aer := p.API.GetChannel(channelId)
  if aer != nil { return aer }
  var tgtPost *model.Post
  if targetPostId != "" {
    tgtPost, aer = p.API.GetPost(targetPostId)
    if aer != nil { return aer }
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
    err = p.movePost(post, tgtChannel, tgtPost)
    if err != nil { return err }
  }
  return nil
}

func (p *Plug) getPostsFromIds(postIds []string) ([]*model.Post, error) {
  posts := make([]*model.Post, 0, len(postIds))
  for _, postId := range(postIds) {
    post, aer := p.API.GetPost(postId)
    if aer != nil {
      return nil, i18n.NewError(i18n.MsgErrorNotExist, "PostId", postId)
    }
    posts = append(posts, post)
  }
  return posts, nil
}

func (p *Plug) assertSourcePermissions(
  userId string, post *model.Post, tgtChannel *model.Channel,
) error {
  p.API.LogDebug("Checking source permissions")

  // Check message delete permission
  var perm *model.Permission
  switch post.UserId {
    case userId: perm = model.PermissionDeletePost
    default: perm = model.PermissionDeleteOthersPosts
  }
  if !p.API.HasPermissionToChannel(userId, post.ChannelId, perm) {
    return i18n.NewError(i18n.MsgErrorPermissionMessage, "PostId", post.Id)
  }

  // Check thread delete permission
  if post.RootId == "" {
    threadPosts, aer := p.getThreadPosts(post.Id)
    if aer != nil { return aer }
    for _, reply := range(threadPosts) {
      var perm *model.Permission
      switch reply.UserId {
        case userId: perm = model.PermissionDeletePost
        default: perm = model.PermissionDeleteOthersPosts
      }
      if !p.API.HasPermissionToChannel(userId, reply.ChannelId, perm) {
        return i18n.NewError(i18n.MsgErrorPermissionReplies, "PostId", post.Id)
      }
    }
  }

  // Check channel restrictions
  if post.ChannelId != tgtChannel.Id {
    srcChannel, aer := p.API.GetChannel(post.ChannelId)
    if aer != nil { return aer }
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
) ([]*model.Post, *model.AppError) {
  thread, aer := p.API.GetPostThread(postId)
  if aer != nil { return nil, aer }
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
  p.API.LogDebug("Checking target permissions")
  if !p.API.HasPermissionToChannel(
    userId, tgtChannel.Id, model.PermissionCreatePost,
  ) {
    name := tgtChannel.DisplayName
    return i18n.NewError(i18n.MsgErrorPermissionTarget, "ChannelName", name)
  }
  return nil
}

func (p *Plug) movePost(
  source *model.Post, channel *model.Channel, root *model.Post,
) error {
  p.API.LogDebug(
    "Moving post", "post", source, "channel", channel, "thread", root,
  )

  // Construct post list
  var posts []*model.Post
  if source.RootId == "" {
    var aer *model.AppError
    posts, aer = p.getThreadPosts(source.Id)
    if aer != nil { return aer }
  } else {
    posts = []*model.Post{ source }
  }
  p.API.LogDebug("Retrieved post list", "posts", posts)

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
      info, aer := p.API.GetFileInfo(fileId)
      if aer != nil { return aer }
      bytes, aer := p.API.GetFile(fileId)
      if aer != nil { return aer }

      // Create new file
      newInfo, aer := p.API.UploadFile(bytes, channel.Id, info.Name)
      if aer != nil { return aer }
      newPost.FileIds = append(newPost.FileIds, newInfo.Id)
    }
    p.API.LogDebug("Copied attachments", "attachments", newPost.FileIds)

    // Create new post
    newPost, aer := p.API.CreatePost(newPost)
    if aer != nil { return aer }
    p.API.LogDebug("Created new post", "post", newPost)

    // Copy reactions
    reactions, aer := p.API.GetReactions(post.Id)
    if aer != nil { return aer }
    for _, reaction := range(reactions) {
      reaction.PostId = newPost.Id
      reaction, aer := p.API.AddReaction(reaction)
      if aer != nil { return aer }
      p.API.LogDebug("Added reaction", "reaction", reaction)
    }

    // Set root post if nil
    if root == nil {
      root = newPost
      p.API.LogDebug("Set post as root for further posts")
    }
  }
  
  // Delete original post
  aer := p.API.DeletePost(source.Id)
  if aer != nil { return aer }
  p.API.LogDebug("Deleted original post")
  return nil
}
