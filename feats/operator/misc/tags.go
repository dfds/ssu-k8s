package misc

import "fmt"

var AllowedTags = map[string]string{
	"dfds.cost.centre":          fmt.Sprintf("%s/dfds.cost.centre", labelPrefix),
	"dfds.service.availability": fmt.Sprintf("%s/dfds.service.availability", labelPrefix),
	"dfds.service.criticality":  fmt.Sprintf("%s/dfds.service.criticality", labelPrefix),
	"dfds.data.classification":  fmt.Sprintf("%s/dfds.data.classification", labelPrefix),
	"dfds.env":                  fmt.Sprintf("%s/dfds.env", labelPrefix),
	"dfds.businessCapability":   fmt.Sprintf("%s/dfds.businessCapability", labelPrefix),
}

func IsTagAllowed(tag string) bool {
	_, ok := AllowedTags[tag]
	return ok
}
