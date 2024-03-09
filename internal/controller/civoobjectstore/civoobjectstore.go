package civoobjectstore

import (
	"context"
	"fmt"
	v1alpha1provider "github.com/crossplane-contrib/provider-civo/apis/civo/provider/v1alpha1"
	"github.com/crossplane-contrib/provider-civo/pkg/civocli"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/providerconfig"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-civo/apis/civo/objectstore/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotCivoObjectStore = "managed resource is not a object store"
	errCreateObjectStore  = "cannot create object store"
	errDeleteObjectStore  = "cannot delete object store"
	errUpdateObjectStore  = "cannot update object store"
	errGetObjectStore     = "cannot get object store"

	objectStoreStatusReady    = "ready"
	objectStoreStatusFailed   = "failed"
	objectStoreStatusCreating = "creating"
)

type connector struct {
	client client.Client
}

type external struct {
	kube       client.Client
	civoClient *civocli.CivoClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CivoObjectStore)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCivoObjectStore)
	}
	civoObjectStore, err := e.civoClient.GetObjectStoreByName(cr.Spec.Name)
	if err != nil || civoObjectStore == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	switch civoObjectStore.Status {
	case objectStoreStatusFailed:
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{ResourceExists: false, ResourceUpToDate: false}, fmt.Errorf("ObjectStore creation failed")

	case objectStoreStatusCreating:
		cr.SetConditions(xpv1.Creating())
		cr.Status.AtProvider.Name = civoObjectStore.Name
		cr.Status.AtProvider.Size = int64(civoObjectStore.MaxSize)
		cr.Status.AtProvider.Status = civoObjectStore.Status
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false}, nil

	case objectStoreStatusReady:
		cr.SetConditions(xpv1.Available())
		cr.Status.AtProvider.Name = civoObjectStore.Name
		cr.Status.AtProvider.Size = int64(civoObjectStore.MaxSize)
		cr.Status.AtProvider.Status = civoObjectStore.Status

		cred := e.civoClient.GetObjectStoreCredential(civoObjectStore.OwnerInfo.CredentialID)
		if cred == nil {
			return managed.ExternalObservation{ResourceExists: false}, errors.New("unable to get object store credentials")
		}

		cd := connectionDetails(civoObjectStore, cred)
		secretName := fmt.Sprintf("%s-%s", cr.Spec.ConnectionDetails.ConnectionSecretNamePrefix, cr.Name)

		connectionSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: cr.Spec.ConnectionDetails.ConnectionSecretNamespace,
			},
			Data: cd,
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
		_, err = e.Update(ctx, mg)
		if err != nil {
			log.Warnf("update error:%s ", err.Error())
			return managed.ExternalObservation{ResourceExists: true}, err
		}
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil

	}

	return managed.ExternalObservation{ResourceExists: false}, nil
}

func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := providerconfig.ControllerName(v1alpha1.CivoObjectStoreGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CivoObjectStoreGroupVersionKind),
		managed.WithExternalConnecter(&connector{client: mgr.GetClient()}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("civoobjectstore", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.CivoObjectStore{}).
		Complete(r)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	objectStore, ok := mg.(*v1alpha1.CivoObjectStore)
	if !ok {
		return nil, errors.New(errNotCivoObjectStore)
	}

	providerConfig := &v1alpha1provider.ProviderConfig{}
	err := c.client.Get(ctx, types.NamespacedName{
		Name: objectStore.Spec.ProviderConfigReference.Name}, providerConfig)
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

func (e *external) Create(_ context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	os, ok := mg.(*v1alpha1.CivoObjectStore)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCivoObjectStore)
	}
	_, err := e.civoClient.CreateObjectStore(os.Spec.Name, os.Spec.Size, os.Spec.AccessKey)
	if err != nil {
		return managed.ExternalCreation{}, errors.New(errCreateObjectStore)
	}

	os.SetConditions(xpv1.Creating())

	return managed.ExternalCreation{
		ExternalNameAssigned: true,
	}, nil
}

func (e external) Update(_ context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	os, ok := mg.(*v1alpha1.CivoObjectStore)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCivoObjectStore)
	}

	objectStore, err := e.civoClient.GetObjectStoreByName(os.Spec.Name)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errGetObjectStore)
	}

	err = e.civoClient.UpdateObjectStore(objectStore.ID, os.Spec.Size)
	if err != nil {
		return managed.ExternalUpdate{}, errors.New(errUpdateObjectStore)
	}
	return managed.ExternalUpdate{}, nil
}

func (e external) Delete(_ context.Context, mg resource.Managed) error {
	os, ok := mg.(*v1alpha1.CivoObjectStore)
	if !ok {
		errors.New(errNotCivoObjectStore)
	}
	objectStore, err := e.civoClient.GetObjectStoreByName(os.Spec.Name)
	if err != nil {
		errors.New(errGetObjectStore)
	}
	os.SetConditions(xpv1.Deleting())

	err = e.civoClient.DeleteObjectStore(objectStore.ID)
	if err != nil {
		return errors.New(errDeleteObjectStore)
	}

	return nil
}