package i18n

import (
  "path/filepath"

  "github.com/mattermost/mattermost-plugin-api/i18n"
)

type PluginAPI = i18n.PluginAPI
type Message = i18n.Message

type I18n struct {
  bundle *i18n.Bundle
  serverLocalizer *Localizer
}

func New(api PluginAPI) (*I18n, error) {
  bundle, err := i18n.InitBundle(api, filepath.Join("assets", "i18n"))
  if err != nil { return nil, err }
  serverLocalizer := &Localizer{ bundle, bundle.GetServerLocalizer() }
  return &I18n{ bundle, serverLocalizer }, nil
}

func (i *I18n) Server() *Localizer {
  return i.serverLocalizer
}

func (i *I18n) User(id string) *Localizer {
  return &Localizer{ i.bundle, i.bundle.GetUserLocalizer(id) }
}

type Localizer struct {
  bundle *i18n.Bundle
  localizer *i18n.Localizer
}

func (l *Localizer) Static(msg *Message) string {
  return l.bundle.LocalizeDefaultMessage(l.localizer, msg)
}

func (l *Localizer) Template(msg *Message, data map[string]string) string {
  return l.bundle.LocalizeWithConfig(l.localizer, &i18n.LocalizeConfig{
    DefaultMessage: msg, TemplateData: data,
  })
}
