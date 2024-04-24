/*
Copyright 2024 The Crossplane Authors.
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

package civovolume

import (
	"context"

	v1alpha1provider "github.com/crossplane-contrib/provider-civo/apis/civo/provider/v1alpha1"
	"github.com/crossplane-contrib/provider-civo/pkg/civocli"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-civo/apis/civo/volume/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/providerconfig"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errGenObservation               = "cannot generate observation"
	errNotCivoVolume                = "managed resource is not a CivoVolume"
	errCreateVolume                 = "cannot create Volume"
	errDeleteVolume                 = "cannot delete Volume"
	volumeStateAvailable            = "available"
	volumeStatePendingInstanceStart = "pending_instance_start"
	volumeStateAttached             = "attached"
)

type connecter struct {
	client client.Client
}

type external struct {
	kube       client.Client
	civoClient *civocli.CivoClient
}

// Setup adds a controller that reconciles ProviderConfigs by accounting for
// their current usage.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.BucketRateLimiter) error {
	name := providerconfig.ControllerName(v1alpha1.CivoVolumeGroupKind)

	o := controller.Options{
		RateLimiter: &rl,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CivoVolumeGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("civovolume", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.CivoVolume{}).
		Complete(r)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	instance, ok := mg.(*v1alpha1.CivoVolume)
	if !ok {
		return nil, errors.New(errNotCivoVolume)
	}

	providerConfig := &v1alpha1provider.ProviderConfig{}

	err := c.client.Get(ctx, types.NamespacedName{
		Name: instance.Spec.ProviderConfigReference.Name}, providerConfig)

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
	cr, ok := mg.(*v1alpha1.CivoVolume)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCivoVolume)
	}
	civoVolume, err := e.civoClient.GetVolume(cr.Spec.Name)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, err
	}
	if civoVolume == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider, err = civocli.GenerateVolumeObservation(civoVolume)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	switch civoVolume.Status {
	case volumeStateAvailable:
		cr.SetConditions(xpv1.Creating())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, nil
	case volumeStateAttached:
		cr.SetConditions(xpv1.Available())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil

	case volumeStatePendingInstanceStart:
		cr.SetConditions(xpv1.Available())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, nil
	}
	return managed.ExternalObservation{}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CivoVolume)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCivoVolume)
	}
	volm, err := e.civoClient.GetVolume(cr.Spec.Name)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	if volm != nil {
		return managed.ExternalCreation{}, nil
	}
	//TODO: Check behavior of when optional fields are not provided
	_, err = e.civoClient.CreateVolume(cr.Spec.Name, cr.Spec.Size, cr.Spec.NetworkID, cr.Spec.ClusterID, cr.Spec.Bootable)
	if err != nil {
		return managed.ExternalCreation{}, errors.New(errCreateVolume)
	}
	cr.SetConditions(xpv1.Creating())
	return managed.ExternalCreation{}, nil

}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CivoVolume)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCivoVolume)
	}

	// Check if the volume needs to be resized.
	if cr.Status.AtProvider.Size != cr.Spec.Size {
		if err := e.civoClient.ResizeVolume(cr.Status.AtProvider.ID, cr.Spec.Size); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "failed to resize volume")
		}
	}

	// Check if the volume needs to be attached to a different instance.
	if cr.Status.AtProvider.InstanceID != cr.Spec.InstanceID && cr.Spec.InstanceID != "" {
		if err := e.civoClient.AttachVolume(cr.Status.AtProvider.ID, cr.Spec.InstanceID); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "failed to attach volume to instance")
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CivoVolume)
	if !ok {
		return errors.New(errNotCivoVolume)
	}
	cr.SetConditions(xpv1.Deleting())
	err := e.civoClient.DeleteVolume(cr.Status.AtProvider.ID)
	return errors.Wrap(err, errDeleteVolume)
}
