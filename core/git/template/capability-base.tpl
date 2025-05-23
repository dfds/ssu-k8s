apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: capability-base-{{index .Vars "capabilityId"}}
  namespace: capability-data
spec:
  releaseName: capability-base-{{index .Vars "capabilityId"}}
  serviceAccountName: flux
  driftDetection: enabled
  chart:
    spec:
      chart: configuration/capability-base
      reconcileStrategy: Revision
      sourceRef:
        kind: GitRepository
        name: ssu-k8s-manifests
        namespace: capability-data
  interval: 1m0s
  install:
    remediation:
      retries: -1
  values:
    capability:
      id: {{index .Vars "capabilityId"}}
      name: {{index .Vars "capabilityName"}}