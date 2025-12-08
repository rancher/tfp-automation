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
	"\n\nAirgap RKE2/K3S Cluster : ~2 hours",
	"\nDualstack RKE2/K3S Cluster : ~5 minutes",
	"\nIPv6 RKE2/K3S Cluster : ~5 minutes",
	"\nNormal RKE2/K3S Cluster : ~5 minutes",
	"\n\nThe cluster access information will be displayed once the setup successfully finishes.",
}

var RancherStageMessage = []string{
	"Please do not close this window while the setup is in progress.",
	"\n\nAirgap Rancher : ~2.5 hours",
	"\nDualstack Rancher : ~20 minutes",
	"\nIPv6 Rancher : ~20 minutes",
	"\nNormal Rancher : ~20 minutes",
	"\nProxy Rancher : ~20 minutes",
	"\nRegistry Rancher : ~2.5 hours",
	"\n\nThe Rancher URL will be displayed once the setup successfully finishes.",
}
