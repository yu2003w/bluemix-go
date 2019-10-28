package main

import (
	"encoding/json"
	"fmt"
	"log"
	gohttp "net/http"
	"os"
	"strconv"
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

type flavor struct {
	Name                      string `json:"name"`
	Provider                  string `json:"provider"`
	Memory                    string `json:"memory"`
	NetworkSpeed              string `json:"networkSpeed"`
	Cores                     string `json:"cores"`
	OS                        string `json:"os"`
	ServerType                string `json:"serverType"`
	Storage                   string `json:"storage"`
	SecondaryStorage          string `json:"secondaryStorage"`
	SecondaryStorageEncrypted bool   `json:"secondaryStorageEncrypted"`
	Deprecated                bool   `json:"deprecated"`
	CorrespondingMachineType  string `json:"correspondingMachineType"`
	IsTrusted                 bool   `json:"isTrusted"`
	Gpus                      string `json:"gpus"`
}

type zone struct {
	ID    string `json:"id"`
	Metro string `json:"metro"`
}

// only pick several fields from flavor
type sflavor struct {
	Name       string `json:"name"`
	OS         string `json:"os"`
	ServerType string `json:"serverType"`
}
type zoneCfg struct {
	ID      string   `json:"id"`
	Metro   string   `json:"metro"`
	Pubvlan []string `json:"public_vlans"`
	Privlan []string `json:"private_vlans"`
	Types   []string `json:"serverTypes"`
}

//ZoneInfo contains zone detailed info and kubernetes versions
type zoneInfo struct {
	K8sversion []string  `json:"k8sVersions"`
	Zones      []zoneCfg `json:"zones"`
}

// GenZoneCfg print zone config information
func GenZoneCfg(apiKey string) {
	trace.Logger = trace.NewLogger("false")
	cli, err := newCli(apiKey)
	if err != nil {
		log.Fatal(err)
		return
	}
	target := v1.ClusterTargetHeader{}
	log.Printf("to get zones")
	zones, err := listZone(cli, target)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("---zones---\n", zones)
	// assembel zonecfg here
	zfg := []zoneCfg{}
	for _, z := range zones {
		fl, err := listFlavor(cli, target, z)
		if err != nil {
			log.Fatal(err)
			return
		}
		//sf := []sflavor{}
		var ts []string
		for _, f := range fl {
			/*
				f := sflavor{
					Name:       f.Name,
					OS:         f.OS,
					ServerType: f.ServerType,
				}
				sf = append(sf, f)
			*/
			ts = append(ts, f.Name)
		}
		vlans, err := listVlan(&v1.CsService{
			Client: cli,
		}, z.ID)
		if err != nil {
			log.Fatal(err)
			return
		}
		var pubv []string
		var priv []string
		for _, l := range vlans {
			val := l.ID + "-" + l.Properties.VlanNumber + "-" + l.Properties.PrimaryRouter
			if l.Type == "public" {

				pubv = append(pubv, val)
			} else if l.Type == "private" {
				priv = append(priv, val)
			} else {
				log.Fatal("unknown vlan type", l.Type)
				return
			}
		}
		zf := zoneCfg{
			ID:      z.ID,
			Metro:   z.Metro,
			Types:   ts,
			Pubvlan: pubv,
			Privlan: priv,
		}
		zfg = append(zfg, zf)
	}

	vers, err := listVersion(&v1.CsService{
		Client: cli,
	})
	var k8svers []string
	for _, v := range vers["kubernetes"] {
		k8svers = append(k8svers, strconv.Itoa(v.Major)+"."+strconv.Itoa(v.Minor)+"."+strconv.Itoa(v.Patch))
	}

	/*
		zfgStr, err := json.MarshalIndent(zfg, "", "  ")
		if err != nil {
			log.Fatal(err)
			return
		}
	*/

	zf := zoneInfo{
		K8sversion: k8svers,
		Zones:      zfg,
	}
	zfstr, err := json.MarshalIndent(zf, "", "  ")
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("Generating iks-zone-cfg")
	fmt.Println(string(zfstr))
}

func writeZfg(zfg []zoneCfg) error {
	f, err := os.OpenFile("iks-zone-cfg.json", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func listZone(cli *client.Client, target v1.ClusterTargetHeader) ([]zone, error) {
	zones := []zone{}
	_, err := cli.Get("/v1/zones", &zones, target.ToMap())
	if err != nil {
		return nil, err
	}
	return zones, nil
}

func listFlavor(cli *client.Client, target v1.ClusterTargetHeader, z zone) ([]flavor, error) {

	fs := []flavor{}
	rawURL := fmt.Sprintf("/v1/datacenters/%s/machine-types", z.ID)
	_, err := cli.Get(rawURL, &fs, target.ToMap())
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func newCli(apiKey string) (*client.Client, error) {
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

	return client.New(config, bluemix.ContainerService, tokenRefreher), nil
}
