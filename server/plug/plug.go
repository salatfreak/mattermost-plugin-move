package plug

import (
  "github.com/mattermost/mattermost-server/v6/plugin"
  pluginapi "github.com/mattermost/mattermost-plugin-api"

  "github.com/salatfreak/mattermost-plugin-move/server/i18n"
)

type Plug struct {
  plugin.MattermostPlugin

  api *pluginapi.Client
  i18n *i18n.I18n
}

func New() *Plug {
  return &Plug{}
}

// Activate plugin
func (p *Plug) OnActivate() error {
  // Create API client
  p.api = pluginapi.NewClient(p.API, p.Driver)

  // Initialize internationalization
  var err error
  p.i18n, err = i18n.New(p.API)
  if err != nil { return err }

  // Create command
  return p.api.SlashCommand.Register(p.createCommand("move"))
}
