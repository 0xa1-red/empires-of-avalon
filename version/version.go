package version

import "fmt"

var (
	Tag       string
	Revision  string
	BuildTime string
)

func Print() {
	fmt.Printf(`Empires of Avalon daemon
-----
Version:   %s
Revision:  %s
Timestamp: %s
`, Tag, Revision, BuildTime)
}
