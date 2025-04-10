package misc

import (
	"fmt"
	"strings"
)

const annotationPrefix = "dfds.cloud"

var LabelFeatureOptInKey = fmt.Sprintf("%s/features", labelPrefix)

func GetFeaturesFromAnnotation(annotations map[string]string) map[string]bool {
	var payload = make(map[string]bool)
	if rawFeatures, ok := annotations[LabelFeatureOptInKey]; ok {
		noWhitespace := strings.Replace(rawFeatures, " ", "", -1)
		feats := strings.Split(noWhitespace, ",")

		for _, feat := range feats {
			payload[feat] = true
		}
	}

	return payload
}
