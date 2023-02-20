package civokubernetes

import (
	"context"
	"fmt"
	"strings"

	"github.com/civo/civogo"
	"github.com/crossplane-contrib/provider-civo/apis/civo/cluster/v1alpha1"
	v1alpha1provider "github.com/crossplane-contrib/provider-civo/apis/civo/provider/v1alpha1"
	"github.com/crossplane-contrib/provider-civo/internal/civoclient"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	deletionMessage = "Cluster is being deleted"
)

type connecter struct {
	client client.Client
}

type external struct {
	kube       client.Client
	civoClient *civogo.Client
}

// Setup sets up a Civo Kubernetes controller.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.CivoKubernetesGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CivoKubernetesGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("civokubernetes", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.CivoKubernetes{}).
		Complete(r)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cluster, ok := mg.(*v1alpha1.CivoKubernetes)
	if !ok {
		return nil, errors.New("managed resource is not a FavouriteDBInstance")
	}

	providerConfig := &v1alpha1provider.ProviderConfig{}

	err := c.client.Get(ctx, types.NamespacedName{
		Name: cluster.Spec.ProviderConfigReference.Name}, providerConfig)

	if err != nil {
		return nil, err
	}

	s := &corev1.Secret{}
	if err := c.client.Get(ctx, types.NamespacedName{Name: providerConfig.Spec.Credentials.SecretRef.Name,
		Namespace: providerConfig.Spec.Credentials.SecretRef.Namespace}, s); err != nil {
		return nil, errors.New("could not find secret")
	}

	civoClient, err := civoclient.NewAPIClient(string(s.Data["credentials"]), providerConfig.Spec.Region)

	if err != nil {
		return nil, err
	}
	return &external{
		kube:       c.client,
		civoClient: civoClient,
	}, nil
}

// nolint
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CivoKubernetes)
	if !ok {
		return managed.ExternalObservation{}, errors.New("invalid object")
	}
	civoCluster, err := e.civoClient.GetKubernetesCluster(meta.GetExternalName(cr.GetObjectMeta()))
	if err != nil {
		return managed.ExternalObservation{}, nil
	}
	if civoCluster == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if strings.Compare(cr.Status.Message, deletionMessage) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	switch civoCluster.Status {
	case "ACTIVE":
		cr.Status.Message = "Cluster is active"
		cd, err := connectionDetails([]byte(civoCluster.KubeConfig), civoCluster.Name)
		if err != nil {
			return managed.ExternalObservation{ResourceExists: true}, err
		}

		// ----------------------------------------------------------------------------
		secretName := fmt.Sprintf("%s-%s", cr.Spec.ConnectionDetails.ConnectionSecretNamePrefix, cr.Name)

		connectionSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: cr.Spec.ConnectionDetails.ConnectionSecretNamespace,
			},
			Data: map[string][]byte{
				"kubeconfig": []byte(civoCluster.KubeConfig),
			},
		}
		err = e.kube.Get(ctx, types.NamespacedName{
			Namespace: cr.Spec.ConnectionDetails.ConnectionSecretNamespace,
			Name:      secretName,
		}, connectionSecret)
		if err != nil {
			err = e.kube.Create(ctx, connectionSecret, &client.CreateOptions{})
			if err != nil {
				return managed.ExternalObservation{ResourceExists: true}, err
			}
		}
		// --------------------------------------------
		_, err = e.Update(ctx, mg)
		if err != nil {
			log.Warnf("update error:%s ", err.Error())
		}
		// --------------------------------------------
		cr.SetConditions(xpv1.Available())
		return managed.ExternalObservation{
			ResourceExists:    true,
			ResourceUpToDate:  true,
			ConnectionDetails: cd,
		}, nil
	case "BUILDING":
		cr.Status.Message = "Cluster is being created"
		cr.SetConditions(xpv1.Creating())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}
	return managed.ExternalObservation{}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CivoKubernetes)
	if !ok {
		return managed.ExternalCreation{}, errors.New("invalid object")
	}

	// at the first call, this id will be the cluster name; civo should return 404
	existingClusterID := meta.GetExternalName(cr.GetObjectMeta())
	civoCluster, err := e.civoClient.GetKubernetesCluster(existingClusterID)
	if err != nil {
		if civogo.DatabaseKubernetesClusterNotFoundError.Is(err) {
			// 404 cluster not found, we continue with the create
		} else {
			// cluster lookup error, return
			return managed.ExternalCreation{}, err
		}
	}
	if civoCluster != nil {
		return managed.ExternalCreation{}, nil
	}
	// Create or Update
	kc := &civogo.KubernetesClusterConfig{
		Name:              cr.Spec.Name,
		Pools:             cr.Spec.Pools,
		Applications:      "",
		CNIPlugin:         *cr.Spec.CNIPlugin,
		KubernetesVersion: *cr.Spec.Version,
		Region:            e.civoClient.Region,
		NetworkID:         *cr.Spec.NetworkID,
	}
	newCluster, err := e.civoClient.NewKubernetesClusters(kc)
	if err != nil {
		log.Warn("Cluster creation failed", err)
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, newCluster.ID)

	cr.SetConditions(xpv1.Creating())

	return managed.ExternalCreation{
		ExternalNameAssigned: true,
	}, nil
}

// nolint
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	desiredCivoCluster, ok := mg.(*v1alpha1.CivoKubernetes)
	if !ok {
		return managed.ExternalUpdate{}, errors.New("invalid object")
	}
	desiredClusterID := meta.GetExternalName(desiredCivoCluster.GetObjectMeta())
	remoteCivoCluster, err := e.civoClient.GetKubernetesCluster(desiredClusterID)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	if remoteCivoCluster == nil {
		return managed.ExternalUpdate{}, nil
	}

	providerConfig := &v1alpha1provider.ProviderConfig{}

	err = e.kube.Get(ctx, types.NamespacedName{
		Name: desiredCivoCluster.Spec.ProviderConfigReference.Name}, providerConfig)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if len(desiredCivoCluster.Spec.Pools) != len(remoteCivoCluster.Pools) || !arePoolsEqual(desiredCivoCluster, remoteCivoCluster) {

		log.Debug("Pools are not equal")
		//TODO: Set region in the civo client once to avoid passing the providerConfig
		desiredClusterConfig := &civogo.KubernetesClusterConfig{
			Pools:  desiredCivoCluster.Spec.Pools,
			Region: e.civoClient.Region, // TODO: Region needs to be on the cluster spec
		}
		if _, err := e.civoClient.UpdateKubernetesCluster(desiredClusterID, desiredClusterConfig); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	if desiredCivoCluster.Spec.Version != nil {
		if *desiredCivoCluster.Spec.Version > remoteCivoCluster.Version {
			log.Info("Updating cluster version")
			desiredClusterConfig := &civogo.KubernetesClusterConfig{
				KubernetesVersion: *desiredCivoCluster.Spec.Version,
				Region:            e.civoClient.Region, // TODO: Region needs to be on the cluster spec
			}
			if _, err := e.civoClient.UpdateKubernetesCluster(desiredClusterID, desiredClusterConfig); err != nil {
				return managed.ExternalUpdate{}, err
			}
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CivoKubernetes)
	if !ok {
		return nil
	}
	civoCluster, err := e.civoClient.GetKubernetesCluster(meta.GetExternalName(cr.GetObjectMeta()))
	if err != nil {
		return err
	}
	if civoCluster == nil {
		log.Warnf("Cluster %s does not exist", civoCluster.Name)
		return nil
	}

	// Removing any existing cluster connection details
	secretName := fmt.Sprintf("%s-%s", cr.Spec.ConnectionDetails.ConnectionSecretNamePrefix, cr.Name)

	connectionSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: cr.Spec.ConnectionDetails.ConnectionSecretNamespace,
		},
	}
	if err := e.kube.Delete(ctx, connectionSecret); err != nil {
		return err
	}
	// ------------------------------------------------
	cr.Status.Message = deletionMessage
	cr.SetConditions(xpv1.Deleting())
	_, err = e.civoClient.DeleteKubernetesCluster(civoCluster.ID)
	return err
}

func arePoolsEqual(desiredCivoCluster *v1alpha1.CivoKubernetes, remoteCivoCluster *civogo.KubernetesCluster) bool {
	for _, desirePool := range desiredCivoCluster.Spec.Pools {
		for _, remotePool := range remoteCivoCluster.Pools {
			if desirePool.ID == remotePool.ID {
				if desirePool.Count != remotePool.Count {
					return false
				}
			}

		}
	}

	return true
}

func connectionDetails(kubeconfig []byte, name string) (managed.ConnectionDetails, error) {
	kcfg, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, errors.New("cannot parse kubeconfig file")

	}
	kctx, ok := kcfg.Contexts[name]
	if !ok {
		return nil, errors.Errorf("context configuration is not found for cluster: %s", name)
	}
	cluster, ok := kcfg.Clusters[kctx.Cluster]
	if !ok {
		return nil, errors.Errorf("cluster configuration is not found: %s", kctx.Cluster)
	}
	auth, ok := kcfg.AuthInfos[kctx.AuthInfo]
	if !ok {
		return nil, errors.Errorf("auth-info configuration is not found: %s", kctx.AuthInfo)
	}

	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey:   []byte(cluster.Server),
		xpv1.ResourceCredentialsSecretCAKey:         cluster.CertificateAuthorityData,
		xpv1.ResourceCredentialsSecretClientCertKey: auth.ClientCertificateData,
		xpv1.ResourceCredentialsSecretClientKeyKey:  auth.ClientKeyData,
		xpv1.ResourceCredentialsSecretKubeconfigKey: kubeconfig,
	}, nil
}
