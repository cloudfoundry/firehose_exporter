package filters

import "strings"

type DeploymentFilter struct {
	deploymentsEnabled map[string]bool
}

func NewDeploymentFilter(filter []string) *DeploymentFilter {
	deploymentsEnabled := make(map[string]bool)

	for _, deploymentName := range filter {
		deploymentsEnabled[strings.Trim(deploymentName, " ")] = true
	}

	return &DeploymentFilter{deploymentsEnabled: deploymentsEnabled}
}

func (f *DeploymentFilter) Enabled(deploymentName string) bool {
	if len(f.deploymentsEnabled) > 0 {
		if f.deploymentsEnabled[deploymentName] {
			return true
		}

		return false
	}

	return true
}
