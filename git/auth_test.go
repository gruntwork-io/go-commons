package git

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
	"github.com/gruntwork-io/terratest/modules/random"
)

func TestIntegrationGitAuth(t *testing.T) {
	t.Parallel()

	tag := "gruntwork-io/go-commons:" + strings.ToLower(random.UniqueId())
	docker.Build(t, "..", &docker.BuildOptions{
		Tags:         []string{tag},
		OtherOptions: []string{"-f", "test/Dockerfile"},
	})
	docker.Run(t, tag, &docker.RunOptions{
		Remove:               true,
		EnvironmentVariables: []string{"GITHUB_OAUTH_TOKEN"},
	})
}
