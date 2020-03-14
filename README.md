# IRC Plugins

This is an assortment of "plugins" for [greboid/irc](https://github.com/greboid/irc/).

## Content

### News

    go run github.com/csmith/ircplugins/cmd/news -host localhost:8000 -token abcdef -channel news

Polls the RSS feeds for a number of news aggregation and reports new items to a channel. 

### Arch

    go run github.com/csmith/ircplugins/cmd/arch -host localhost:8000 -token abcdef

Responds to `!arch <package>` with search results from both the main arch repos and the AUR.
