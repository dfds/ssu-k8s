apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: capability-base-{{index .Vars "capabilityId"}}
  namespace: capability-data
spec:
  releaseName: capability-base-{{index .Vars "capabilityId"}}
  chart:
    spec:
      chart: configuration/capability-base
      reconcileStrategy: Revision
      sourceRef:
        kind: GitRepository
        name: capability-data
        namespace: capability-data
  interval: 1m0s
  install:
    remediation:
      retries: 3
  values:
    capability:
      id: {{index .Vars "capabilityId"}}
      name: {{index .Vars "capabilityName"}}