package version

// This variable is set at build time using -ldflags parameters. For example, we typically set this flag in circle.yml
// to the latest Git tag when building our Go apps:
//
// build-go-binaries --app-name my-app --dest-path bin --ld-flags "-X github.com/gruntwork-io/go-commons/version.VERSION=$CIRCLE_TAG"
// go build -o my-app -ldflags "-X github.com/gruntwork-io/go-commons/version.VERSION=v1.0.0"
//
// For more info, see: http://stackoverflow.com/a/11355611/483528

var (
	VERSION = "latest"
)

// Version composes a version of the package
func Version() string {
	return VERSION
}
