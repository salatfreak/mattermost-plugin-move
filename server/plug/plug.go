package plug

import (
  "github.com/mattermost/mattermost-server/v6/plugin"

  "github.com/salatfreak/mattermost-plugin-move/server/i18n"
)

type Plug struct {
  plugin.MattermostPlugin

  i18n *i18n.I18n
}

func New() *Plug {
  return &Plug{}
}

// Activate plugin
func (p *Plug) OnActivate() error {
  // Initialize internationalization
  var err error
  p.i18n, err = i18n.New(p.API)
  if err != nil { return err }

  // Create command
  aer := p.API.RegisterCommand(p.createCommand("move"))
  if aer != nil { return aer }

  return nil
}
