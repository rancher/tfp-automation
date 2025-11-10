module github.com/rancher/tfp-automation

go 1.24.2

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.6.27 // for compatibilty with docker 20.10.x
	github.com/docker/distribution => github.com/docker/distribution v2.8.2+incompatible // rancher-machine requires a replace is set
	github.com/docker/docker => github.com/docker/docker v20.10.27+incompatible // rancher-machine requires a replace is set
	github.com/docker/go-connections => github.com/docker/go-connections v0.4.0
	github.com/rancher/rancher/pkg/apis => github.com/rancher/rancher/pkg/apis v0.0.0-20251029210211-fd1057ed2062
	github.com/rancher/rancher/pkg/client => github.com/rancher/rancher/pkg/client v0.0.0-20251029210211-fd1057ed2062

	go.qase.io/client => github.com/rancher/qase-go/client v0.0.0-20250627195016-142ff3dfec16
	go.qase.io/qase-api-client => github.com/qase-tms/qase-go/qase-api-client v1.2.0
	k8s.io/api => k8s.io/api v0.34.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.34.1
	k8s.io/apiserver => k8s.io/apiserver v0.34.1
	k8s.io/client-go => k8s.io/client-go v0.34.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.34.1
	k8s.io/component-base => k8s.io/component-base v0.34.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.34.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.34.1
	k8s.io/cri-api => k8s.io/cri-api v0.34.1
	k8s.io/cri-client => k8s.io/cri-client v0.34.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.34.1
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.34.1
	k8s.io/endpointslice => k8s.io/endpointslice v0.34.1
	k8s.io/externaljwt => k8s.io/externaljwt v0.0.0-20240209024834-5f1e9e5f2a0c
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.34.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.34.1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20250710124328-f3f2b991d03b
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.34.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.34.1
	k8s.io/kubectl => k8s.io/kubectl v0.34.1
	k8s.io/kubelet => k8s.io/kubelet v0.34.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.34.1
	k8s.io/metrics => k8s.io/metrics v0.34.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.34.1
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.34.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.34.1
)

require (
	github.com/stretchr/testify v1.11.1
	k8s.io/api v0.34.1
	k8s.io/apimachinery v0.34.1
	k8s.io/apiserver v0.34.1 // indirect
)

require (
	github.com/gruntwork-io/terratest v0.49.0
	github.com/imdario/mergo v0.3.16
	github.com/rancher/norman v0.7.1
	github.com/rancher/rancher/pkg/apis v0.0.0
	github.com/rancher/shepherd v0.0.0-20251106204618-0f330e53bbc0
	github.com/rancher/tests v0.0.0-20251110212725-072c07d47741
	github.com/rancher/tests/actions v0.0.0-20251110212725-072c07d47741
	github.com/sirupsen/logrus v1.9.3
	go.qase.io/qase-api-client v0.0.0-00010101000000-000000000000
)

require (
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
)

require (
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/hashicorp/go-getter/v2 v2.2.3 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kubereboot/kured v1.13.1 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/pkg/sftp v1.13.5 // indirect
	github.com/rancher/ali-operator v0.0.5 // indirect
	github.com/rancher/wrangler/v3 v3.3.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/tools v0.37.0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/validator.v2 v2.0.1 // indirect
	k8s.io/component-helpers v0.34.1 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
)

require (
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/aws/aws-sdk-go v1.55.6 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/hcl/v2 v2.22.0
	github.com/hashicorp/terraform-json v0.23.0 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-zglob v0.0.2-0.20190814121620-e3c945676326 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/tmccombs/hcl2json v0.6.4 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/zclconf/go-cty v1.15.0
	golang.org/x/crypto v0.43.0
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/oauth2 v0.32.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/creasty/defaults v1.5.2 // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/evanphx/json-patch v5.9.11+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.72.0 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rancher/aks-operator v1.13.0-rc.1 // indirect
	github.com/rancher/apiserver v0.7.0 // indirect
	github.com/rancher/eks-operator v1.13.0-rc.1 // indirect
	github.com/rancher/fleet/pkg/apis v0.14.0-rc.1 // indirect
	github.com/rancher/gke-operator v1.13.0-rc.1 // indirect
	github.com/rancher/lasso v0.2.5 // indirect
	github.com/rancher/rke v1.8.0-rc.4 // indirect
	github.com/rancher/system-upgrade-controller/pkg/apis v0.0.0-20250710162344-185ff9f785cd // indirect
	github.com/rancher/wrangler v1.1.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/term v0.36.0 // indirect
	golang.org/x/time v0.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apiextensions-apiserver v0.34.1 // indirect
	k8s.io/cli-runtime v0.34.1 // indirect
	k8s.io/client-go v12.0.0+incompatible // indirect
	k8s.io/component-base v0.34.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-aggregator v0.34.1 // indirect
	k8s.io/kube-openapi v0.31.5 // indirect
	k8s.io/kubectl v0.34.1 // indirect
	k8s.io/kubernetes v1.34.1 // indirect
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397 // indirect
	sigs.k8s.io/cli-utils v0.37.2 // indirect
	sigs.k8s.io/cluster-api v1.10.6 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/kustomize/api v0.20.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.20.1 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
