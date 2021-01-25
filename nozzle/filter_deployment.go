package nozzle

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type FilterDeployment struct {
	deployments []string
}

func NewFilterDeployment(deployments ...string) *FilterDeployment {
	return &FilterDeployment{deployments: deployments}
}

func (f *FilterDeployment) IsFiltered(envelope *loggregator_v2.Envelope) bool {
	if len(f.deployments) == 0 {
		return false
	}
	currentDepl, ok := envelope.GetTags()["deployment"]
	if !ok {
		return false
	}
	for _, deployment := range f.deployments {
		if currentDepl == deployment {
			return false
		}
	}
	return true
}

func (f *FilterDeployment) SetDeployments(deployments ...string) {
	f.deployments = deployments
}
