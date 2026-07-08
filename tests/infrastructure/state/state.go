package state

import "sync"

type appState struct {
	StageMsg string
	ErrorMsg string
	Mutex    sync.Mutex
}

var State = &appState{}

var ClusterStageMessage = []string{
	"Please do not close this window while the setup is in progress.",
	"\n\nAirgap RKE2/K3S Cluster : ~45 minutes",
	"\nDualstack RKE2/K3S Cluster : ~5 minutes",
	"\nIPv6 RKE2/K3S Cluster : ~5 minutes",
	"\nNormal RKE2/K3S Cluster : ~5 minutes",
	"\nProxy RKE2/K3S Cluster : ~5 minutes",
	"\n\nThe cluster access information will be displayed once the setup successfully finishes.",
}

var RancherStageMessage = []string{
	"Please do not close this window while the setup is in progress.",
	"\n\nAirgap Rancher : ~1 hour",
	"\nDualstack Rancher : ~20 minutes",
	"\nIPv6 Rancher : ~20 minutes",
	"\nNormal Rancher : ~20 minutes",
	"\nProxy Rancher : ~20 minutes",
	"\nRegistry Rancher : ~2.5 hours",
	"\n\nThe Rancher URL will be displayed once the setup successfully finishes.",
}

var RegistryStageMessage = []string{
	"Please do not close this window while the setup is in progress.",
	"\n\nAll Registries : ~3 hours",
	"\nAuthenticated Registry : ~1 hour",
	"\nNon-Authenticated Registry : ~40 minutes",
	"\nECR : ~50 minutes",
	"\n\nThe registry access information will be displayed once the setup successfully finishes.",
}
