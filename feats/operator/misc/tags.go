package misc

import "fmt"

var AllowedTags = map[string]string{
	"dfds.cost.centre": fmt.Sprintf("dfds.cost.centre"),
	//"dfds.owner":              fmt.Sprintf("%s/owner", labelPrefix),
	"dfds.service.availability": fmt.Sprintf("dfds.service.availability"),
	"dfds.service.criticality":  fmt.Sprintf("dfds.service.criticality"),
	"dfds.data.classification":  fmt.Sprintf("dfds.data.classification"),
	"dfds.env":                  fmt.Sprintf("dfds.env"),
	"dfds.businessCapability":   fmt.Sprintf("dfds.businessCapability"),
}

func IsTagAllowed(tag string) bool {
	_, ok := AllowedTags[tag]
	return ok
}
