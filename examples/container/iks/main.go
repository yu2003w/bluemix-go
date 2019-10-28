package main

import (
	"fmt"
	"log"
	gohttp "net/http"
	"os"
	"time"

	"github.com/IBM-Cloud/bluemix-go"
	v1 "github.com/IBM-Cloud/bluemix-go/api/container/containerv1"
	"github.com/IBM-Cloud/bluemix-go/authentication"
	"github.com/IBM-Cloud/bluemix-go/client"
	"github.com/IBM-Cloud/bluemix-go/endpoints"
	"github.com/IBM-Cloud/bluemix-go/helpers"
	"github.com/IBM-Cloud/bluemix-go/http"
	"github.com/IBM-Cloud/bluemix-go/rest"
	"github.com/IBM-Cloud/bluemix-go/trace"
)

func main() {
	trace.Logger = trace.NewLogger("false")

	apiKey := ""
	if len(os.Args) != 2 {
		log.Println(os.Args)
		log.Fatal("missed parameters")
		return
	}

	if os.Args[1] == "zonecfg" {
		GenZoneCfg(apiKey)
		return
	} else if os.Args[1] != "test" {
		log.Fatal("unknown arguments", os.Args)
		return
	}

	cli, err := NewCli(apiKey)
	if err != nil {
		log.Fatal(err)
		return
	}

	vers, err := listVersion(cli)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("---versions---\n", vers)

	out, err := listClusters(cli)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("---clusters---\n", out)

	vl, err := listVlan(cli, "dal10")
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("---vlans---\n", vl)

	id, err := createCluster(cli)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("---cluster created---", id)

	fmt.Println("---waiting for cluster ready---")
	time.Sleep(time.Minute * 10)
	if err = deleteCluster(cli, id); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("---cluster deleted---", id)

}

func NewCli(apiKey string) (v1.ContainerServiceAPI, error) {
	config := &bluemix.Config{
		MaxRetries:      helpers.Int(3),
		BluemixAPIKey:   apiKey,
		HTTPTimeout:     180 * time.Second,
		RetryDelay:      helpers.Duration(30 * time.Second),
		EndpointLocator: endpoints.NewEndpointLocator(""),
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

	log.Println("Access Token:", config.IAMAccessToken)
	log.Println("Refresh Token:", config.IAMRefreshToken)

	return &v1.CsService{
		Client: client.New(config, bluemix.ContainerService, tokenRefreher),
	}, nil
}

func listClusters(cli v1.ContainerServiceAPI) ([]v1.ClusterInfo, error) {
	clustersAPI := cli.Clusters()
	target := v1.ClusterTargetHeader{}
	out, err := clustersAPI.List(target)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return out, nil
}

func listVersion(cli v1.ContainerServiceAPI) (v1.V1Version, error) {
	clustersAPI := cli.KubeVersions()
	target := v1.ClusterTargetHeader{}
	out, err := clustersAPI.ListV1(target)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return out, nil
}

func listVlan(cli v1.ContainerServiceAPI, zone string) ([]v1.DCVlan, error) {
	vlanAPI := cli.Vlans()
	target := v1.ClusterTargetHeader{}
	out, err := vlanAPI.List(zone, target)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return out, nil
}

func createCluster(cli v1.ContainerServiceAPI) (string, error) {
	var clusterInfo = v1.ClusterCreateRequest{
		Name:        "cluster_test_jared",
		Datacenter:  "dal12",
		MachineType: "u2c.2x4",
		WorkerNum:   1,
		PrivateVlan: "",
		PublicVlan:  "",
		Isolation:   "public",
	}
	clusterAPI := cli.Clusters()
	target := v1.ClusterTargetHeader{}
	out, err := clusterAPI.Create(clusterInfo, target)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return out.ID, nil
}

func deleteCluster(cli v1.ContainerServiceAPI, id string) error {
	clusterAPI := cli.Clusters()
	target := v1.ClusterTargetHeader{}
	err := clusterAPI.Delete(id, target)

	return err
}
