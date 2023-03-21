package civokubernetes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/civo/civogo"
	"github.com/crossplane-contrib/provider-civo/apis/civo/cluster/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	cptest "github.com/crossplane/crossplane-runtime/pkg/test"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testClusterExternalID = "69a23478-a89e-41d2-97b1-6f4c341cee70"

func getFakeClusterv1Alpha1() *v1alpha1.CivoKubernetes {
	networkId := "fake-network-id"
	firewallId := "fake-firewall-id"
	version := "1.22.2-k3s1"
	cnipluginflannel := "flannel"

	return &v1alpha1.CivoKubernetes{
		TypeMeta: v1.TypeMeta{
			Kind:       v1alpha1.CivoKubernetesKind,
			APIVersion: v1alpha1.CivoKubernetesKindAPIVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-cluster",
		},
		Spec: v1alpha1.CivoKubernetesSpec{
			ResourceSpec: getFakeResourceSpec(),
			Name:         "test-cluster",
			Region:       "LON1",
			NetworkID:    &networkId,
			FirewallID:   &firewallId,
			Version:      &version,
			CNIPlugin:    &cnipluginflannel, //required
			Pools: []civogo.KubernetesClusterPoolConfig{
				getFakePool(1),
			},
			Applications: []string{"kubernetes-dashboard"},
		},
	}
}

func getFakePool(id int) civogo.KubernetesClusterPoolConfig {
	return civogo.KubernetesClusterPoolConfig{
		ID:    fmt.Sprintf("test-pool-%d", id),
		Count: 1,
		Size:  "g3.k3s.small",
	}
}

func getFakeResourceSpec() xpv1.ResourceSpec {
	return xpv1.ResourceSpec{
		ProviderConfigReference: &xpv1.Reference{
			Name: "test-provider",
		},
		WriteConnectionSecretToReference: &xpv1.SecretReference{
			Name: "connection-secret",
		},
		DeletionPolicy: xpv1.DeletionOrphan,
	}
}

func TestUpdate(t *testing.T) {
	testCasesConfig := []ConfigErrorClientForTesting{
		{
			Method: "GET",
			Value: []ValueErrorClientForTesting{
				{
					URL: fmt.Sprintf("/v2/kubernetes/clusters/%s", testClusterExternalID),
					ResponseBody: fmt.Sprintf(`{
					"id": "%s",
					"name": "test-cluster",
					"version": "1.22.2-k3s1",
					"cluster_type": "k3s",
					"status": "ACTIVE",
					"ready": true,
					"num_target_nodes": 1,
					"target_nodes_size": "g3.k3s.small",
					"built_at": "0001-01-01T00:00:00Z",
					"kubernetes_version": "1.23.6-k3s1",
					"created_at": "2023-03-01T19:24:36Z",
					"required_pools": [
						{
						"id": "node-pool",
						"size": "g3.k3s.small",
						"count": 1
						}
					],
					"firewall_id": "test-firewall-id",
					"master_ipv6": "",
					"applications": "kubernetes-dashboard",
					"network_id": "test-network-id",
					"pools": [{
						"id": "node-pool",
						"size": "g3.k3s.small",
						"count": 1
					}]
				}`, testClusterExternalID),
				},
			},
		},
		{
			Method: "PUT",
			Value: []ValueErrorClientForTesting{
				{
					URL:         "/v2/kubernetes/clusters/",
					RequestBody: `{"region":"TEST","pools":[{"id":"test-pool-1","count":1,"size":"g3.k3s.small"},{"id":"test-pool-2","count":1,"size":"g3.k3s.small"}]}`,
					ResponseBody: fmt.Sprintf(`{
						"id": "%s"
					}`, testClusterExternalID),
				},
				{
					URL:         "/v2/kubernetes/clusters/",
					RequestBody: `{"region":"TEST","instance_firewall":"fake-firewall-id"}`,
					ResponseBody: fmt.Sprintf(`{
						"id": "%s"
					}`, testClusterExternalID),
				},
				{
					URL:         "/v2/kubernetes/clusters/",
					RequestBody: `{"region":"TEST","tags":"lets-add-a-test-tag"}`,
					ResponseBody: fmt.Sprintf(`{
						"id": "%s"
					}`, testClusterExternalID),
				},
				{
					URL:         "/v2/kubernetes/clusters/",
					RequestBody: `{"region":"TEST","applications":"cilium istio"}`,
					ResponseBody: fmt.Sprintf(`{
						"id": "%s"
					}`, testClusterExternalID),
				},
			},
		},
	}
	civoClient, server, err, results := NewErrorClientForTesting(testCasesConfig)
	if err != nil {
		t.Error(err)
	}
	defer server.Close()

	mockExternal := &external{
		kube:       cptest.NewMockClient(),
		civoClient: civoClient,
	}

	ctx := context.Background()
	cr := getFakeClusterv1Alpha1()
	meta.SetExternalName(cr, testClusterExternalID) // since this is an update, we already have an external ID
	cr.Spec.Pools = append(cr.Spec.Pools, getFakePool(2))
	cr.Spec.Applications = []string{"cilium", "istio"} // TODO: this should remove kubernetes-dashboard
	cr.Spec.Tags = []string{"lets-add-a-test-tag"}

	_, errU := mockExternal.Update(ctx, cr)
	if errU != nil {
		t.Error(errU)
	}

	if results.Completed[0].URL.Path != "/v2/kubernetes/clusters/"+testClusterExternalID {
		// make sure that the function called the GET url
		t.Log("Test did not call GET url properly")
		t.FailNow()
	}

	for _, req := range results.Failed {
		body, _ := io.ReadAll(req.Body)
		t.Errorf("Could not match request %s %s\nBody: \"%s\"",
			req.Method,
			req.URL.String(),
			body,
		)
	}
}

func TestCreate(t *testing.T) {
	t.Helper()

	testCasesConfig := []ConfigErrorClientForTesting{
		{
			Method: "GET",
			Value: []ValueErrorClientForTesting{
				{
					URL:          "/v2/kubernetes/clusters",
					StatusCode:   404,
					ResponseBody: `{"code":"database_kubernetes_cluster_not_found"}`,
				},
			},
		},
		{
			Method: "POST",
			Value: []ValueErrorClientForTesting{
				{
					URL:        "/v2/kubernetes/clusters",
					StatusCode: 200,
					ResponseBody: fmt.Sprintf(`{
						"id": "%s",
						"name": "your-cluster-name",
						"version": "2",
						"status": "ACTIVE",
						"ready": true,
						"num_target_nodes": 1,
						"target_nodes_size": "g2.xsmall",
						"built_at": "2019-09-23T13:04:23.000+01:00",
						"kubeconfig": "YAML_VERSION_OF_KUBECONFIG_HERE\n",
						"kubernetes_version": "0.8.1",
						"api_endpoint": "https://your.cluster.ip.address:6443",
						"master_ip": "your.cluster.ip.address",
						"dns_entry": "69a23478-a89e-41d2-97b1-6f4c341cee70.k8s.civo.com",
						"tags": [],
						"created_at": "2019-09-23T13:02:59.000+01:00",
						"firewall_id": "test-firewall-id",
						"cni_plugin": "flannel"
					}`, testClusterExternalID),
				},
			},
		},
	}

	civoClient, server, _, results := NewErrorClientForTesting(testCasesConfig)
	defer server.Close()

	mockExternal := &external{
		kube:       cptest.NewMockClient(),
		civoClient: civoClient,
	}

	ctx := context.Background()
	cr := getFakeClusterv1Alpha1()
	createOp, err := mockExternal.Create(ctx, cr)
	if err != nil {
		t.Error(err)
	}

	if createOp.ExternalNameAssigned == false {
		t.Error("ExternalName was not assigned")
	}

	if meta.GetExternalName(cr) != testClusterExternalID {
		t.Errorf("Wrong external name, expected %s, found %s", testClusterExternalID, meta.GetExternalName(cr))
	}

	for _, req := range results.Failed {
		body, _ := io.ReadAll(req.Body)
		t.Errorf("Could not match request %s %s\nBody: \"%s\"",
			req.Method,
			req.URL.String(),
			body,
		)
	}
}

type ConfigErrorClientForTesting struct {
	Method string
	Value  []ValueErrorClientForTesting
}

type ValueErrorClientForTesting struct {
	RequestBody  string
	URL          string
	ResponseBody string
	Handler      func(rw http.ResponseWriter, req *http.Request)
	StatusCode   int
}

type errorClientResults struct {
	Completed []*http.Request
	Failed    []*http.Request
}

// NewAdvancedClientForTesting initializes a Client connecting to a local test server
// it allows for specifying methods and records and returns all requests made
func NewErrorClientForTesting(responses []ConfigErrorClientForTesting) (*civogo.Client, *httptest.Server, error, *errorClientResults) {
	results := &errorClientResults{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var responseSent bool

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			return
		}

		req.Body = io.NopCloser(bytes.NewBuffer(body))

		for _, criteria := range responses {
			// we check the HTTP method first
			if req.Method != criteria.Method {
				continue
			}

			for _, criteria := range criteria.Value {
				if !strings.HasPrefix(req.URL.Path, criteria.URL) {
					// simple match on request body by prefix
					continue
				}

				if !strings.HasPrefix(string(body), criteria.RequestBody) {
					// match on request body by prefix, so we can pass "" to match all
					continue
				}

				responseSent = true

				if criteria.StatusCode > 0 {
					rw.WriteHeader(criteria.StatusCode)
				}
				rw.Write([]byte(criteria.ResponseBody))
			}
		}

		if responseSent {
			results.Completed = append(results.Completed, req)
		} else {
			results.Failed = append(results.Failed, req)
			fmt.Println("failed to find a matching request")
			fmt.Printf("%s %s\n", req.Method, req.URL.String())
			fmt.Println("Request Body: ", string(body))
			rw.Write([]byte(`{"result": "failed to find a matching request handler"}`))
		}
	}))

	client, err := civogo.NewClientForTestingWithServer(server)

	return client, server, err, results
}
