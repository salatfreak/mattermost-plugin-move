package args

import (
  "strings"
  "regexp"

  "github.com/mattermost/mattermost-server/v6/model"

  "github.com/salatfreak/mattermost-plugin-move/server/i18n"
)

var (
  msgIDExp = regexp.MustCompile(`(?m)^[a-z0-9]+$`)
  msgURLExp = regexp.MustCompile(
    `(?m)^(https?://[-_.a-z0-9]+(?::\d+)?)/(?:[-a-z0-9]+)/pl/([a-z0-9]+)`,
  )
  chanNameExp = regexp.MustCompile(`(?m)^[-_a-z0-9]+$`)
  chanURLExp = regexp.MustCompile(
    `(?m)^(https?://[-_.a-z0-9]+(?::\d+)?)` +
    `/(?:[-a-z0-9]+)/channels/([-_a-z0-9]+)$`,
  )
)

func Parse(args *model.CommandArgs) ([]string, error) {
  // Get source list
  cmdWords := strings.Split(args.Command, " ")[1:]
  sources := make([]string, 0, len(cmdWords))
  for _, word := range(cmdWords) {
    if word != "" { sources = append(sources, word) }
  }
  if len(sources) == 0 { return nil, i18n.NewError(i18n.MsgErrorNoMessages) }

  // Extract message IDs from sources
  for i, source := range(sources) {
    if match := msgURLExp.FindStringSubmatch(source); len(match) > 0 {
      if match[1] != args.SiteURL {
        return nil, i18n.NewError(i18n.MsgErrorOtherInstance)
      }
      sources[i] = match[2]
    } else if !msgIDExp.MatchString(source) {
      return nil, i18n.NewError(i18n.MsgErrorNotAMessage, "PostId", source)
    }
  }

  // Return sources
  return sources, nil
}
