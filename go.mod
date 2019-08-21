module github.com/mattermost/mattermost-plugin-starter-template

go 1.12

require (
	github.com/blang/semver v3.6.1+incompatible // indirect
	github.com/mattermost/mattermost-server v0.0.0-20190810005745-1dcfaf34d0b6
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.3.0
)

// Workaround for https://github.com/golang/go/issues/30831 and fallout.
replace github.com/golang/lint => github.com/golang/lint v0.0.0-20190227174305-8f45f776aaf1
