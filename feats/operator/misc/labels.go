package misc

import "fmt"

const labelPrefix = "dfds.cloud"

var LabelCapabilityKey = fmt.Sprintf("%s/capability", labelPrefix)
var LabelReconcileKey = fmt.Sprintf("%s/reconcile", labelPrefix)
