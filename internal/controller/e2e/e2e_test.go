package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/civo/civogo"
	"github.com/crossplane-contrib/provider-civo/apis"
	"github.com/crossplane-contrib/provider-civo/internal/controller/civoinstance"
	"github.com/crossplane-contrib/provider-civo/internal/controller/civokubernetes"
	civoprovider "github.com/crossplane-contrib/provider-civo/internal/controller/provider"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/joho/godotenv"
	"gopkg.in/alecthomas/kingpin.v2"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var CivoRegion, CivoURL string

var e2eTest *E2ETest

var (
	app   = kingpin.New(filepath.Base(os.Args[0]), "Template support for Crossplane.").DefaultEnvars()
	debug = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
)

type E2ETest struct {
	civo         *civogo.Client
	cluster      *civogo.KubernetesCluster
	tenantClient client.Client
}

// TestMain provisions and then cleans up a cluster for testing against
func TestMain(m *testing.M) {
	e2eTest = &E2ETest{}

	// Recover from a panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
		// Ensure that we clean up the cluster after test tests have run
		e2eTest.cleanUpCluster()
	}()

	// Recover from a SIGINT
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == syscall.SIGINT {
				e2eTest.cleanUpCluster()
			}
		}
	}()

	// Load .env from the project root
	godotenv.Load("../.env")

	// Provision a new cluster
	e2eTest.provisionCluster()
	defer e2eTest.cleanUpCluster()

	// 2. Wait for the cluster to be provisioned
	retry(30, 10*time.Second, func() error {
		cluster, err := e2eTest.civo.GetKubernetesCluster(e2eTest.cluster.ID)
		if err != nil {
			return err
		}
		if cluster.Status != "ACTIVE" {
			return fmt.Errorf("Cluster is not available yet: %s", cluster.Status)
		}
		return nil
	})

	var err error
	e2eTest.cluster, err = e2eTest.civo.GetKubernetesCluster(e2eTest.cluster.ID)
	if err != nil {
		log.Panicf("Unable to fetch ACTIVE Cluster: %s", err.Error())
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(e2eTest.cluster.KubeConfig))
	if err != nil {
		log.Panic(err)
	}

	// Connect to the cluster
	cl, err := client.New(config, client.Options{})
	if err != nil {
		log.Panic(err)
	}
	e2eTest.tenantClient = cl

	secret := &corev1.Secret{}
	err = cl.Get(context.Background(), client.ObjectKey{Namespace: "kube-system", Name: "civo-api-access"}, secret)
	if err != nil {
		log.Panicf("Unable get civo-api-access secret: %s", err.Error())
	}

	// Run the provider
	go run(secret, e2eTest.cluster.KubeConfig)
	time.Sleep(1 * time.Minute)

	// Run the tests
	fmt.Println("Running Tests")
	exitCode := m.Run()

	e2eTest.cleanUpCluster()

	os.Exit(exitCode)

}

func (e *E2ETest) provisionCluster() {
	APIKey := os.Getenv("CIVO_API_KEY")
	if APIKey == "" {
		log.Panic("CIVO_API_KEY env variable not provided")
	}

	CivoRegion = os.Getenv("CIVO_REGION")
	if CivoRegion == "" {
		CivoRegion = "LON1"
	}

	CivoURL := os.Getenv("CIVO_API_URL")
	if CivoURL == "" {
		CivoURL = "https://api.civo.com"
	}

	var err error
	e.civo, err = civogo.NewClientWithURL(APIKey, CivoURL, CivoRegion)
	if err != nil {
		log.Panicf("Unable to initialise Civo Client: %s", err.Error())
	}

	// List Clusters
	list, err := e.civo.ListKubernetesClusters()
	if err != nil {
		log.Panicf("Unable to list Clusters: %s", err.Error())
	}
	for _, cluster := range list.Items {
		if cluster.Name == "ccm-e2e-test" {
			e.cluster = &cluster
			return
		}
	}

	// List Networks
	network, err := e.civo.GetDefaultNetwork()
	if err != nil {
		log.Panicf("Unable to get Default Network: %s", err.Error())
	}

	clusterConfig := &civogo.KubernetesClusterConfig{
		Name:      "ccm-e2e-test",
		Region:    CivoRegion,
		NetworkID: network.ID,
		Pools: []civogo.KubernetesClusterPoolConfig{
			{
				Count: 2,
				Size:  "g4s.kube.xsmall",
			},
		},
	}

	e.cluster, err = e.civo.NewKubernetesClusters(clusterConfig)
	if err != nil {
		log.Panicf("Unable to provision new cluster: %s", err.Error())
	}
}

func (e *E2ETest) cleanUpCluster() {
	fmt.Println("Attempting Test Cleanup")
	if e.cluster != nil {
		fmt.Printf("Deleting Cluster: %s\n", e.cluster.ID)
		e.civo.DeleteKubernetesCluster(e.cluster.ID)
	}
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Println("retrying after error:", err)
			time.Sleep(sleep)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func run(secret *corev1.Secret, kubeconfig string) {
	err := os.WriteFile("./kubeconfig", []byte(kubeconfig), 0644)
	if err != nil {
		panic(err)
	}

	// Read env var from in cluster secret
	APIURL := string(secret.Data["api-url"])
	APIKey := string(secret.Data["api-key"])
	Region := string(secret.Data["region"])
	ClusterID := string(secret.Data["cluster-id"])

	if APIURL == "" || APIKey == "" || Region == "" || ClusterID == "" {
		fmt.Println("CIVO_API_URL, CIVO_API_KEY, CIVO_REGION, CIVO_CLUSTER_ID environment variables must be set")
		os.Exit(1)
	}

	klog.Infof("Starting Civo Crossplane Provider with CIVO_API_URL: %s, CIVO_REGION: %s, CIVO_CLUSTER_ID: %s", APIURL, Region, ClusterID)
	klog.Info("Please make sure CRD's are installed in the cluster. They are inside /package/crds/")

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-template"))
	if *debug {
		// The controller-runtime runs with a no-op logger by default. It is
		// *very* verbose even at info level, so we only provide it a real
		// logger when we're running in debug mode.
		ctrl.SetLogger(zl)
	}

	cfg, err := ctrl.GetConfig()
	kingpin.FatalIfError(err, "Cannot get API server rest config")

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		LeaderElection:   *app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").OverrideDefaultFromEnvar("LEADER_ELECTION").Bool(),
		LeaderElectionID: "crossplane-leader-election-provider-template",
		SyncPeriod:       app.Flag("sync", "Controller manager sync period such as 300ms, 1.5h, or 2h45m").Short('s').Default("1h").Duration(),
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")

	rl := ratelimiter.NewDefaultProviderRateLimiter(ratelimiter.DefaultProviderRPS)
	kingpin.FatalIfError(apis.AddToScheme(mgr.GetScheme()), "Cannot add Template APIs to scheme")
	kingpin.FatalIfError(civokubernetes.Setup(mgr, log, rl), "Cannot setup Civo K3 Cluster controllers")
	kingpin.FatalIfError(civoinstance.Setup(mgr, log, rl), "Cannot setup Civo Instance controllers")
	kingpin.FatalIfError(civoprovider.Setup(mgr, log, rl), "Cannot setup Provider controllers")
	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}
