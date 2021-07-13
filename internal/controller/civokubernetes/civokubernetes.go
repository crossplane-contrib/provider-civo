package CivoKubernetes

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/civo/civo-crossplane-provider/apis/civo/cluster/v1alpha1"
	v1alpha1provider "github.com/civo/civo-crossplane-provider/apis/civo/provider/v1alpha1"
	"github.com/civo/civo-crossplane-provider/pkg/civocli"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
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
	civoClient *civocli.CivoClient
}

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

	civoClient, err := civocli.NewCivoClient(string(s.Data["credentials"]), providerConfig.Spec.Region)

	if err != nil {
		return nil, err
	}
	return &external{
		kube:       c.client,
		civoClient: civoClient,
	}, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CivoKubernetes)
	if !ok {
		return managed.ExternalObservation{}, errors.New("invalid object")
	}
	civoCluster, err := e.civoClient.GetK3sCluster(cr.Spec.Name)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, err
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
	civoCluster, err := e.civoClient.GetK3sCluster(cr.Spec.Name)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	if civoCluster != nil {
		return managed.ExternalCreation{}, nil
	}
	// Create or Update
	err = e.civoClient.CreateNewK3sCluster(cr.Spec.Name, cr.Spec.Instances, cr.Spec.Size, cr.Spec.Applications)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	cr.SetConditions(xpv1.Creating())

	return managed.ExternalCreation{
		ExternalNameAssigned: true,
	}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CivoKubernetes)
	if !ok {
		return nil
	}
	civoCluster, err := e.civoClient.GetK3sCluster(cr.Spec.Name)
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
	e.kube.Delete(ctx, connectionSecret)
	// ------------------------------------------------
	cr.Status.Message = deletionMessage
	cr.SetConditions(xpv1.Deleting())
	return e.civoClient.DeleteK3sCluster(civoCluster.Name)
}

func connectionDetails(kubeconfig []byte, name string) (managed.ConnectionDetails, error) {
	kcfg, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, errors.New("cannot parse kubeconfig file")

	}
	kctx, ok := kcfg.Contexts[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("context configuration is not found for cluster: %s", name))
	}
	cluster, ok := kcfg.Clusters[kctx.Cluster]
	if !ok {
		return nil, errors.New(fmt.Sprintf("cluster configuration is not found: %s", kctx.Cluster))
	}
	auth, ok := kcfg.AuthInfos[kctx.AuthInfo]
	if !ok {
		return nil, errors.New(fmt.Sprintf("auth-info configuration is not found: %s", kctx.AuthInfo))
	}

	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey:   []byte(cluster.Server),
		xpv1.ResourceCredentialsSecretCAKey:         cluster.CertificateAuthorityData,
		xpv1.ResourceCredentialsSecretClientCertKey: auth.ClientCertificateData,
		xpv1.ResourceCredentialsSecretClientKeyKey:  auth.ClientKeyData,
		xpv1.ResourceCredentialsSecretKubeconfigKey: kubeconfig,
	}, nil
}
