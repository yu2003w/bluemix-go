package containerv2

import (
	"fmt"

	"github.com/IBM-Cloud/bluemix-go/client"
)

//Worker ...
type Worker struct {
	Billing           string `json:"billing,omitempty"`
	Flavor            string `json:"flavor"`
	ID                string `json:"id"`
	KubeVersion       KubeDetails
	Location          string          `json:"location"`
	PoolID            string          `json:"poolid"`
	PoolName          string          `json:"poolName"`
	LifeCycle         WorkerLifeCycle `json:"lifecycle"`
	Health            HealthStatus    `json:"health"`
	NetworkInterfaces []Network       `json:"networkInterfaces"`
}

type KubeDetails struct {
	Actual    string `json:"actual"`
	Desired   string `json:"desired"`
	Eos       string `json:"eos"`
	MasterEOS string `json:"masterEos"`
	Target    string `json:"target"`
}
type HealthStatus struct {
	Message string `json:"message"`
	State   string `json:"state"`
}
type WorkerLifeCycle struct {
	ReasonForDelete    string `json:"reasonForDelete"`
	ActualState        string `json:"actualState"`
	DesiredState       string `json:"desiredState"`
	Message            string `json:"message"`
	MessageDate        string `json:"messageDate"`
	MessageDetails     string `json:"messageDetails"`
	MessageDetailsDate string `json:"messageDetailsDate"`
	PendingOperation   string `json:"pendingOperation"`
}

type Network struct {
	Cidr      string `json:"cidr"`
	IpAddress string `json:"ipAddress"`
	Primary   bool   `json:"primary"`
	SubnetID  string `json:"subnetID"`
}

//Workers ...
type Workers interface {
	ListByWorkerPool(clusterIDOrName, workerPoolIDOrName string, showDeleted bool, target ClusterTargetHeader) ([]Worker, error)
	ListWorkers(clusterIDOrName string, showDeleted bool, target ClusterTargetHeader) ([]Worker, error)
}

type worker struct {
	client *client.Client
}

func newWorkerAPI(c *client.Client) Workers {
	return &worker{
		client: c,
	}
}

//ListByWorkerPool ...
func (r *worker) ListByWorkerPool(clusterIDOrName, workerPoolIDOrName string, showDeleted bool, target ClusterTargetHeader) ([]Worker, error) {
	rawURL := fmt.Sprintf("/v2/vpc/getWorkers?cluster=%s&showDeleted=%t", clusterIDOrName, showDeleted)
	if len(workerPoolIDOrName) > 0 {
		rawURL += "&pool=" + workerPoolIDOrName
	}
	workers := []Worker{}
	_, err := r.client.Get(rawURL, &workers, target.ToMap())
	if err != nil {
		return nil, err
	}
	return workers, err
}

//ListWorkers ...
func (r *worker) ListWorkers(clusterIDOrName string, showDeleted bool, target ClusterTargetHeader) ([]Worker, error) {
	rawURL := fmt.Sprintf("/v2/vpc/getWorkers?cluster=%s&showDeleted=%t", clusterIDOrName, showDeleted)
	workers := []Worker{}
	_, err := r.client.Get(rawURL, &workers, target.ToMap())
	if err != nil {
		return nil, err
	}
	return workers, err
}
