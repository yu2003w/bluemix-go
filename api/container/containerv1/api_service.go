package containerv1

import (
	gohttp "net/http"

	bluemix "github.com/IBM-Cloud/bluemix-go"
	"github.com/IBM-Cloud/bluemix-go/authentication"
	"github.com/IBM-Cloud/bluemix-go/client"
	"github.com/IBM-Cloud/bluemix-go/http"
	"github.com/IBM-Cloud/bluemix-go/rest"
	"github.com/IBM-Cloud/bluemix-go/session"
)

//ErrCodeAPICreation ...
const ErrCodeAPICreation = "APICreationError"

//ContainerServiceAPI is the Aramda K8s client ...
type ContainerServiceAPI interface {
	Albs() Albs
	Clusters() Clusters
	Workers() Workers
	WorkerPools() WorkerPool
	WebHooks() Webhooks
	Subnets() Subnets
	KubeVersions() KubeVersions
	Vlans() Vlans
}

//ContainerService holds the client
type CsService struct {
	*client.Client
}

//New ...
func New(sess *session.Session) (ContainerServiceAPI, error) {
	config := sess.Config.Copy()
	err := config.ValidateConfigForService(bluemix.ContainerService)
	if err != nil {
		return nil, err
	}
	if config.HTTPClient == nil {
		config.HTTPClient = http.NewHTTPClient(config)
	}
	tokenRefreher, err := authentication.NewIAMAuthRepository(config, &rest.Client{
		DefaultHeader: gohttp.Header{
			"User-Agent": []string{http.UserAgent()},
		},
		HTTPClient: config.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	if config.IAMAccessToken == "" {
		err := authentication.PopulateTokens(tokenRefreher, config)
		if err != nil {
			return nil, err
		}
	}
	if config.Endpoint == nil {
		ep, err := config.EndpointLocator.ContainerEndpoint()
		if err != nil {
			return nil, err
		}
		config.Endpoint = &ep
	}

	return &CsService{
		Client: client.New(config, bluemix.ContainerService, tokenRefreher),
	}, nil
}

//Albs implement albs API
func (c *CsService) Albs() Albs {
	return newAlbAPI(c.Client)
}

//Clusters implements Clusters API
func (c *CsService) Clusters() Clusters {
	return newClusterAPI(c.Client)
}

//Workers implements Cluster Workers API
func (c *CsService) Workers() Workers {
	return newWorkerAPI(c.Client)
}

//WorkerPools implements Cluster WorkerPools API
func (c *CsService) WorkerPools() WorkerPool {
	return newWorkerPoolAPI(c.Client)
}

//Subnets implements Cluster Subnets API
func (c *CsService) Subnets() Subnets {
	return newSubnetAPI(c.Client)
}

//Webhooks implements Cluster WebHooks API
func (c *CsService) WebHooks() Webhooks {
	return newWebhookAPI(c.Client)
}

//KubeVersions implements Cluster WebHooks API
func (c *CsService) KubeVersions() KubeVersions {
	return newKubeVersionAPI(c.Client)
}

//Vlans implements DC Cluster Vlan API
func (c *CsService) Vlans() Vlans {
	return newVlanAPI(c.Client)
}
