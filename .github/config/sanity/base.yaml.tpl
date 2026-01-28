rancher:
  host: "${HOSTNAME_PREFIX}.${AWS_ROUTE_53_ZONE}"
  adminPassword: "${RANCHER_ADMIN_PASSWORD}"
  insecure: true
  cleanup: true

terraform:
  cni: "${CNI}"
  defaultClusterRoleForProjectMembers: "true"
  downstreamClusterProvider: "${DOWNSTREAM_PROVIDER}"
  enableNetworkPolicy: false
  provider: "${PROVIDER}"
  privateKeyPath: "${SSH_PRIVATE_KEY_PATH}"
  resourcePrefix: "${HOSTNAME_PREFIX}"
  windowsPrivateKeyPath: "${WINDOWS_SSH_PRIVATE_KEY_PATH}"

  privateRegistries:
    url: "${PRIVATE_REGISTRY_URL}"
    username: "${DOCKERHUB_USERNAME}"
    password: "${DOCKERHUB_PASSWORD}"
    insecure: true
    authConfigSecretName: "${AUTH_CONFIG_SECRET_NAME}"
    mirrorHostname: "${PRIVATE_REGISTRY_MIRROR_HOSTNAME}"
    mirrorEndpoint: "${PRIVATE_REGISTRY_MIRROR_ENDPOINT}"

standalone:
  bootstrapPassword: "${RANCHER_ADMIN_PASSWORD}"
  certManagerVersion: "${CERT_MANAGER_VERSION}"
  certType: "${CERT_TYPE}"
  chartVersion: "${RANCHER_CHART_VERSION}"
  osUser: "${OS_USER}"
  osGroup: "${OS_GROUP}"
  rancherChartRepository: "${RANCHER_HELM_CHART_URL}"
  rancherHostname: "${HOSTNAME_PREFIX}.${AWS_ROUTE_53_ZONE}"
  rancherImage: "${RANCHER_IMAGE}"
  rancherTagVersion: "${RANCHER_VERSION}"
  registryUsername: "${PRIVATE_REGISTRY_USERNAME}"
  registryPassword: "${PRIVATE_REGISTRY_PASSWORD}"
  repo: "${RANCHER_REPO}"

terratest:
  pathToRepo: "${PATH_TO_REPO}"
  etcdCount: 3
  controlPlaneCount: 2
  workerCount: 3
  windowsNodeCount: 1