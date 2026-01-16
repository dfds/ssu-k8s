package misc

import "fmt"

var AllowedTags = map[string]string{
	"dfds.cost.centre": fmt.Sprintf("%s/cost-centre", labelPrefix),
	//"dfds.owner":                fmt.Sprintf("%s/owner", labelPrefix),
	"dfds.service.availability": fmt.Sprintf("%s/service-availability", labelPrefix),
	"dfds.service.criticality":  fmt.Sprintf("%s/service-criticality", labelPrefix),
	"dfds.data.classification":  fmt.Sprintf("%s/data-classification", labelPrefix),
}

func IsTagAllowed(tag string) bool {
	_, ok := AllowedTags[tag]
	return ok
}
