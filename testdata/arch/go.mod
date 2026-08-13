module example.com/arch

go 1.25.0

require go.wdy.de/nago v0.0.0

require (
	github.com/worldiety/i18n v0.0.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)

replace go.wdy.de/nago => ../nago

replace github.com/worldiety/i18n => ../i18n
