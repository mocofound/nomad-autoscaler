package nomad

import (
	"fmt"

	nomadHelper "github.com/hashicorp/nomad-autoscaler/helper/nomad"
	"github.com/hashicorp/nomad-autoscaler/strategy"
	targetpkg "github.com/hashicorp/nomad-autoscaler/target"
	"github.com/hashicorp/nomad/api"
)

type Target struct {
	client *api.Client
}

func (t *Target) SetConfig(config map[string]string) error {

	cfg := nomadHelper.ConfigFromMap(config)

	client, err := api.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to instantiate Nomad client: %v", err)
	}
	t.client = client

	return nil
}

func (t *Target) Count(config map[string]string) (int64, error) {
	subTarget, err := t.subTarget(config)
	if err != nil {
		return 0, err
	}
	return subTarget.Count(config)
}

func (t *Target) Scale(action strategy.Action, config map[string]string) error {
	subTarget, err := t.subTarget(config)
	if err != nil {
		return err
	}
	return subTarget.Scale(action, config)
}

func (t *Target) subTarget(config map[string]string) (targetpkg.Target, error) {
	var target targetpkg.Target

	switch config["property"] {
	case "count":
		target = &NomadGroupCount{client: t.client}
	default:
		return nil, fmt.Errorf("invalid property %s", config["property"])

	}
	return target, nil
}
