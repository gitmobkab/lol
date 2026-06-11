package data

import "fmt"

var (
	Version string = "dev"
	BuildTime string = "local"
	CommitHash string = "dirty"
)

/* 
returns a formatted version string with build information,
the format is:

v<Version> (<OS>/<ARCH>)
*/
func GetVersion() string {
	return fmt.Sprintf("v%s (%s/%s)", Version, OS, ARCH)
}

/* 
returns a more detailed formatted version string with build information,
the format is:

v<Version> (<OS>/<ARCH>) [<BuildTime>] *<CommitHash>
*/
func GetDetailedVersion() string {
	return fmt.Sprintf("v%s (%s/%s) [%s] *%s", Version, OS, ARCH, BuildTime, CommitHash)
}