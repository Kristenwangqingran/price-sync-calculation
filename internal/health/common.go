package health

import (
	"context"

	"git.garena.com/shopee/common/spkit"
	"git.garena.com/shopee/common/spkit/pkg/spex"
	"git.garena.com/shopee/common/spkit/runtime"
	"git.garena.com/shopee/common/ulog"
	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/logging"
	"git.garena.com/shopee/platform/golang_splib/sps"
)

var (
	agent sps.Agent
)

func createAndRegisterSpsAgent() (sps.Agent, error) {
	instanceID, configKey, err := generateNewSpexAgentCredentials()
	if err != nil {
		return nil, err
	}

	agent, err = sps.NewAgent(sps.WithInstanceID(instanceID), sps.WithConfigKey(configKey))
	if err != nil {
		logging.GetLogger(context.Background()).Error("failed to initialize sps agent for health check / smoke test",
			ulog.String("instance_id", instanceID), ulog.Error(err))
		return nil, err
	}

	if err = agent.Register(context.Background()); err != nil {
		logging.GetLogger(context.Background()).Error("failed to register sps agent for health check / smoke test",
			ulog.String("instance_id", instanceID), ulog.Error(err))
		return nil, err
	}

	return agent, nil
}

func generateNewSpexAgentCredentials() (string, string, error) {
	sduID, err := spex.SduID()
	if err != nil {
		logging.GetLogger(context.Background()).Error("failed to get spex sdu id for health check / smoke test",
			ulog.String("sdu_id", sduID), ulog.Error(err))
		return "", "", err
	}

	instanceID, err := sps.GenerateInstanceID(spkit.DefaultService().Name(), "", runtime.Env(), "", sduID, "")
	if err != nil {
		logging.GetLogger(context.Background()).Error("failed to generate spex InstanceID for health check / smoke test",
			ulog.String("instance_id", instanceID),
			ulog.String("service_name", spkit.DefaultService().Name()),
			ulog.String("env", runtime.Env()),
			ulog.String("sdu_id", sduID),
			ulog.Error(err))
		return "", "", err
	}

	configKey, err := spex.ConfigKey()
	if err != nil {
		logging.GetLogger(context.Background()).Error("failed to get spex configKey for health check / smoke test", ulog.Error(err))
		return "", "", err
	}

	return instanceID, configKey, nil
}
