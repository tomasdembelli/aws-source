package integration

import (
	"fmt"
	"github.com/overmindtech/sdp-go"
)

func getInstanceID(sdpInstances []*sdp.Item) (string, error) {
	if len(sdpInstances) != 1 {
		return "", fmt.Errorf("expected 1 instance, got %v", len(sdpInstances))
	}

	instanceIDAttrVal, err := sdpInstances[0].GetAttributes().Get("instanceId")
	if err != nil {
		return "", fmt.Errorf("failed to get instanceId: %v", err)
	}

	instanceID := instanceIDAttrVal.(string)
	if instanceID == "" {
		return "", fmt.Errorf("instanceId is empty")
	}

	return instanceID, nil
}
