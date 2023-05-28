# Mattermost Move Plugin
Platforms like [mattermost][mattermost] allow for well organized communication
with messages being thematically grouped in channels and threads. In many cases
topics still get mixed up and messages end up in the wrong place. This plugin
lets you move messages and threads between channels and attach them to other
threads.

The plugin is based on the [Plugin Starter Template][template] and heavily
inspired by the [Wrangler Plugin][wrangler]. It aims to be more powerful yet
more simple and intuitive than the Wrangler but also lacks a couple of its
features.

This plugin is in beta status. Feel free to test it out and offer feedback.

[mattermost]: https://mattermost.com/
[template]: https://github.com/mattermost/mattermost-plugin-starter-template
[wrangler]: https://github.com/gabrieljackson/mattermost-plugin-wrangler

## Major differences to the Wrangler plugin
1. Messages can be moved out of threads and into other threads.
2. Attaching a message that started a thread to another thread, will move all
   of its replies into that thread as well instead of deleting them.
3. Message timestamps are preserved, so they might be mixed into the ordering
   of the existing messages of the target channel or thread.
4. Moving permissions are based on the ability to create and delete messages.
   You may only move a message or thread to another channel if you are allowed
   to delete all messages that are to be moved and to create messages in the
   target channel. Additionally you are not allowed to move messages between
   teams or out of private channels.

## User interface
There is only one slash command and no graphical user interface but the command
is dead simple: Simply type `/move` into the channel or thread where the
messages are supposed to go and append a space separated list of message
links. You can retrieve a messages link by hovering the messages, clicking the
"⋯" icon and then "Copy Link".

![Demo](https://salatfreak.github.io/images/mattermost-plugin-move.gif)

## Installation
1. Download the latest release from the [release page][releases]
2. In Mattermost navigate to "System Console" -> "Plugins" ->
   "Plugin Management"
3. Click "Chose File" and upload the archive

[releases]: https://github.com/salatfreak/mattermost-plugin-move/releases

## Development
The plugin can be built by installing [make][make] and [go][go] and running
`make` in the root directory of the repository. 

For testing the plugin locally, you can start a preview server as a podman or
docker container.

```sh
podman run --rm \
  --name mattermost \
  --publish 8065:8065 \
  docker.io/mattermost/mattermost-preview
```

Building the plugin and publishing it to the local (or a remote) server can be
accomplished by setting the following environment variables accordingly and
running `make deploy`.

```sh
export MM_SERVICESETTINGS_ENABLEDEVELOPER=true
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=<ADMIN USERNAME>
export MM_ADMIN_PASSWORD=<ADMIN PASSWORD>
```

Building and deploying also works flawlessly from alpine based podman or docker
containers.

[make]: https://www.gnu.org/software/make/
[go]: https://go.dev/

## Open tasks
- Write unit tests
- Add more translations (can easily be done via [json files][i18n])

[i18n]: https://github.com/salatfreak/mattermost-plugin-move/tree/master/assets/i18n
