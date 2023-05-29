package plug

import (
  "errors"

  "github.com/mattermost/mattermost-server/v6/model"
  "github.com/mattermost/mattermost-server/v6/plugin"

  "github.com/salatfreak/mattermost-plugin-move/server/i18n"
  "github.com/salatfreak/mattermost-plugin-move/server/args"
)

func (p *Plug) createCommand(trigger string) *model.Command {
  return &model.Command{
    Trigger: trigger,
    AutoComplete: true,
    AutoCompleteHint: p.i18n.Server().Static(i18n.MsgCommandHint),
    AutoCompleteDesc: p.i18n.Server().Static(i18n.MsgCommandDesc),
  }
}

func (p *Plug) ExecuteCommand(
  c *plugin.Context, cmd *model.CommandArgs,
) (*model.CommandResponse, *model.AppError) {
  // Parse args
  sourceIds, err := args.Parse(cmd)
  if err != nil {
    return p.responseFromError(err, p.i18n.User(cmd.UserId)), nil
  }

  // Move messages
  err = p.runMoveMessages(
    cmd.TeamId, cmd.ChannelId, cmd.RootId, cmd.UserId, sourceIds,
  )
  if err != nil {
    return p.responseFromError(err, p.i18n.User(cmd.UserId)), nil
  }

  // Return successfully
  p.api.Log.Debug("Messages moved successfully")
  return &model.CommandResponse{}, nil
}

func (p *Plug) responseFromError(
  err error, localizer *i18n.Localizer,
) *model.CommandResponse {
  var userError *i18n.Error
  if errors.As(err, &userError) {
    // Return localized message for user errors
    p.api.Log.Debug("Message moving user error", "error", userError)
    return &model.CommandResponse{
      ResponseType: model.CommandResponseTypeEphemeral,
      Text: userError.Localize(localizer),
    }
  } else {
    // Return generic message for server errors
    p.api.Log.Error("Message moving server error", "error", err.Error())
    return &model.CommandResponse{
      ResponseType: model.CommandResponseTypeEphemeral,
      Text: localizer.Static(i18n.MsgErrorServer),
    }
  }
}
