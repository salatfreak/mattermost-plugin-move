package main

import (
  "github.com/mattermost/mattermost-server/v6/plugin"

  "github.com/salatfreak/mattermost-plugin-move/server/plug"
)

func main() {
  plugin.ClientMain(plug.New())
}
