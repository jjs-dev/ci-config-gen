package main

import (
	"fmt"
	"strings"

	"github.com/jjs-dev/ci-config-gen/actions"
)

func generatePublishImageScript(config ciConfig) string {
	lines := make([]string, 0)
	lines = append(lines, "set -euxo pipefail")
	header := `
# GENERATED FILE DO NOT EDIT
if [ "$GITHUB_REF" = "refs/heads/master" ]
then
  TAG="latest"
elif [ "$GITHUB_REF" = "refs/heads/trying" ]
then
  TAG="dev"
else
  echo "unknown GITHUB_REF: $GITHUB_REF"
  exit 1
fi
echo ${{ secrets.GHCR_TOKEN }} | docker login ghcr.io -u $GITHUB_ACTOR --password-stdin`
	lines = append(lines, header)

	for _, image := range config.DockerImages {
		lines = append(lines, fmt.Sprintf("docker tag %s ghcr.io/jjs-dev/%s:$TAG", image, image))
		lines = append(lines, fmt.Sprintf("docker push ghcr.io/jjs-dev/%s:$TAG", image))
	}

	return strings.Join(lines, "\n")
}

func makePublishWorkflow(root string, config ciConfig) actions.Workflow {
	publishJob := actions.Job{
		RunsOn: actions.UbuntuRunner,
		If:     "github.event_name == 'push'",
		Steps: []actions.Step{
			makeCheckoutStep(),
			{
				Name: "Build artifacts",
				Run:  "bash ci/publish-build.sh",
			},
			{
				Name: "Publish docker images",
				Run:  "bash ci/publish-images.sh",
			},
		},
	}

	w := actions.Workflow{
		Name: "publish",
		On: actions.Trigger{
			PullRequest: actions.EmptyStruct{},
			Push: actions.PushTrigger{
				Branches: []string{"staging", "trying", "master"},
			},
		},
		Jobs: map[string]actions.Job{
			"publish": publishJob,
		},
	}

	return w
}
