package i18n

var (
  MsgCommandHint = &Message{
    ID: "command.hint",
    Other: "[messages...]",
  }
  MsgCommandDesc = &Message{
    ID: "command.desc",
    Other: "Move messages (IDs or URLs) to current channel or thread",
  }
  MsgErrorServer = &Message{
    ID: "error.server",
    Other: "A server error occured.",
  }
  MsgErrorNoMessages = &Message{
    ID: "error.no_messages",
    Other: "You didn't specify any messages.",
  }
  MsgErrorOtherInstance = &Message{
    ID: "error.other_instance",
    Other: "Cannot move messages from other Mattermost.",
  }
  MsgErrorNotAMessage = &Message{
    ID: "error.not_a_message",
    Other: "{{.PostId}} is not a message ID or URL.",
  }
  MsgErrorAttachItself = &Message{
    ID: "error.attach_itself",
    Other: "Can't attach message to itself.",
  }
  MsgErrorNewerMessage = &Message{
    ID: "error.newer_message",
    Other: "Can't attach to a newer message.",
  }
  MsgErrorNotExist = &Message{
    ID: "error.not_exist",
    Other: "Message {{.PostId}} doesn't exist.",
  }
  MsgErrorPermissionMessage = &Message{
    ID: "error.permission_message",
    Other: "You are not allowed to move message {{.PostId}}.",
  }
  MsgErrorPermissionReplies = &Message{
    ID: "error.permission_replies",
    Other: "You are not allowed to move all replies in message {{.PostId}}.",
  }
  MsgErrorOtherTeam = &Message{
    ID: "error.other_team",
    Other: "Can't move messages between teams.",
  }
  MsgErrorPrivateChannel = &Message{
    ID: "error.private_channel",
    Other: "Can't move messages out of private channels.",
  }
  MsgErrorPermissionTarget = &Message{
    ID: "error.permission_target",
    Other: "You are not allowed to create messages in channel {{.ChannelName}}.",
  }
)
