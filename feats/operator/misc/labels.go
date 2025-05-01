package misc

import (
	"fmt"
	"strings"
)

const labelPrefix = "dfds.cloud"

var LabelCapabilityKey = fmt.Sprintf("%s/capability", labelPrefix)
var LabelReconcileKey = fmt.Sprintf("%s/reconcile", labelPrefix)
var LabelTypeKey = fmt.Sprintf("%s/type", labelPrefix)
var LabelAwsAccountKey = fmt.Sprintf("%s/aws-account", labelPrefix)
var LabelContextIdKey = fmt.Sprintf("%s/context-id", labelPrefix)

func GetFeaturesFromLabel(labels map[string]string) map[string]bool {
	var payload = make(map[string]bool)
	if rawFeatures, ok := labels[LabelFeatureOptInKey]; ok {
		noWhitespace := strings.Replace(rawFeatures, " ", "", -1)
		strings.Split(noWhitespace, ",")
	}

	return payload
}
