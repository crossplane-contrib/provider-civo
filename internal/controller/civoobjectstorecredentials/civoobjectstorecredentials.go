/*
Copyright 2020 The Crossplane Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package civoobjectstorecredentials

import (
	"context"
	"fmt"

	"github.com/apex/log"
	"github.com/civo/civogo"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/providerconfig"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-civo/apis/civo/objectstorecredential/v1alpha1"
	v1alpha1provider "github.com/crossplane-contrib/provider-civo/apis/civo/provider/v1alpha1"
	"github.com/crossplane-contrib/provider-civo/pkg/civocli"
)

const (
	errNotCivoObjectStoreCredentials = "managed resource is not a CivoObjectStoreCredentials"
	errDeleteObjectStoreCredentials  = "cannot delete ObjectStoreCredentials"
)

type connecter struct {
	client client.Client
}

type external struct {
	kube         client.Client
	civoGoClient *civogo.Client
}

// Setup adds a controller that reconciles ProviderConfigs by accounting for
// their current usage.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := providerconfig.ControllerName(v1alpha1.CivoObjectStoreCredentialGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CivoObjectStoreCredentialGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("civoobjectstorecredentials", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.CivoObjectStoreCredential{}).
		Complete(r)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	objectStoreCredentials, ok := mg.(*v1alpha1.CivoObjectStoreCredential)
	if !ok {
		return nil, errors.New(errNotCivoObjectStoreCredentials)
	}

	providerConfig := &v1alpha1provider.ProviderConfig{}

	err := c.client.Get(ctx, types.NamespacedName{
		Name: objectStoreCredentials.Spec.ProviderConfigReference.Name}, providerConfig)

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
		kube:         c.client,
		civoGoClient: civoClient.CivoGoClient,
	}, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CivoObjectStoreCredential)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCivoObjectStoreCredentials)
	}
	civoObjectStoreCredentials := FindObjectStoreCredential(e.civoGoClient, cr.Spec.Name)
	if civoObjectStoreCredentials == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	switch civoObjectStoreCredentials.Status {
	case "ready":
		cr.Status.AtProvider.ID = civoObjectStoreCredentials.ID
		cr.Status.AtProvider.Status = civoObjectStoreCredentials.Status

		cd := connectionDetails(civoObjectStoreCredentials)
		secretName := fmt.Sprintf("%s-%s", cr.Spec.ConnectionDetails.ConnectionSecretNamePrefix, cr.Name)

		connectionSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: cr.Spec.ConnectionDetails.ConnectionSecretNamespace,
			},
			Data: cd,
		}
		err := e.kube.Get(ctx, types.NamespacedName{
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
		cr.SetConditions(xpv1.Available())
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil

	case "creating":
		cr.Status.Message = "ObjectStoreCredentials is being created"
		cr.SetConditions(xpv1.Creating())
		cr.Status.AtProvider.ID = civoObjectStoreCredentials.ID
		cr.Status.AtProvider.Status = civoObjectStoreCredentials.Status
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, nil

	case "failed":
		cr.Status.Message = "ObjectStoreCredentials creation failed"
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, fmt.Errorf("ObjectStoreCredentials creation failed")
	}
	return managed.ExternalObservation{ResourceExists: false}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CivoObjectStoreCredential)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCivoObjectStoreCredentials)
	}

	civoObjectStoreCredentials := FindObjectStoreCredential(e.civoGoClient, cr.Spec.Name)
	if civoObjectStoreCredentials != nil {
		return managed.ExternalCreation{}, nil
	}

	createObjectStoreCredentialsRequest := civogo.CreateObjectStoreCredentialRequest{
		Name:              cr.Spec.Name,
		Region:            e.civoGoClient.Region,
		AccessKeyID:       nil,
		SecretAccessKeyID: nil,
	}
	if cr.Spec.AccessKeyID != nil {
		createObjectStoreCredentialsRequest.AccessKeyID = cr.Spec.AccessKeyID
	}

	_, err := e.civoGoClient.NewObjectStoreCredential(&createObjectStoreCredentialsRequest)
	if err != nil {
		return managed.ExternalCreation{
			ExternalNameAssigned: true,
		}, err
	}
	cr.SetConditions(xpv1.Creating())

	return managed.ExternalCreation{
		ExternalNameAssigned: true,
	}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CivoObjectStoreCredential)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCivoObjectStoreCredentials)
	}
	civoObjectStoreCredentials := FindObjectStoreCredential(e.civoGoClient, cr.Spec.Name)
	var updateObjectStoreCredentialsRequest civogo.UpdateObjectStoreCredentialRequest
	if cr.Spec.AccessKeyID == nil || cr.Spec.AccessKeyID != &civoObjectStoreCredentials.AccessKeyID {
		updateObjectStoreCredentialsRequest = civogo.UpdateObjectStoreCredentialRequest{
			AccessKeyID: cr.Spec.AccessKeyID,
			Region:      e.civoGoClient.Region,
		}
		_, err := e.civoGoClient.UpdateObjectStoreCredential(civoObjectStoreCredentials.ID, &updateObjectStoreCredentialsRequest)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CivoObjectStoreCredential)
	if !ok {
		return errors.New(errNotCivoObjectStoreCredentials)
	}
	objectStoreCredentials := FindObjectStoreCredential(e.civoGoClient, cr.Spec.Name)
	cr.SetConditions(xpv1.Deleting())
	_, err := e.civoGoClient.DeleteObjectStoreCredential(objectStoreCredentials.ID)
	return errors.Wrap(err, errDeleteObjectStoreCredentials)
}
