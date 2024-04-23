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

package civonetwork

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

	"github.com/crossplane-contrib/provider-civo/apis/civo/network/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/providerconfig"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotCivoNetwork = "managed resource is not a CivoNetwork"
	errDeleteNetwork  = "cannot delete network"
	errUpdateNetwork  = "cannot update network"
	errGenObservation = "cannot generate observation"
)

var (
	// NetworkActive defines active Status of a network
	NetworkActive = "Active"

	// NetworkDeleting defines deleting Status of a network
	NetworkDeleting = "Deleting"
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
	name := providerconfig.ControllerName(v1alpha1.CivoNetworkGroupKind)

	o := controller.Options{
		RateLimiter: &rl,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CivoNetworkGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("civonetwork", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.CivoNetwork{}).
		Complete(r)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	network, ok := mg.(*v1alpha1.CivoNetwork)
	if !ok {
		return nil, errors.New(errNotCivoNetwork)
	}

	providerConfig := &v1alpha1provider.ProviderConfig{}

	err := c.client.Get(ctx, types.NamespacedName{
		Name: network.Spec.ProviderConfigReference.Name}, providerConfig)

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
	cr, ok := mg.(*v1alpha1.CivoNetwork)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCivoNetwork)
	}
	civoNetwork, err := e.civoClient.GetNetwork(cr.Spec.Name)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, err
	}
	if civoNetwork == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider, err = civocli.GenerateNetworkObservation(civoNetwork)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	switch civoNetwork.Status {
	case NetworkActive:
		// Check if NameserversV4 field has changed
		if !stringSlicesEqual(civoNetwork.NameserversV4, cr.Spec.NameserversV4) {
			cr.SetConditions(xpv1.Creating())
			return managed.ExternalObservation{
				ResourceExists:   true,
				ResourceUpToDate: false,
			}, nil
		}
		cr.SetConditions(xpv1.Available())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	case NetworkDeleting:
		cr.SetConditions(xpv1.Deleting())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, nil
	}
	return managed.ExternalObservation{}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CivoNetwork)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCivoNetwork)
	}
	civoNetwork, err := e.civoClient.GetNetwork(cr.Spec.Name)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	if civoNetwork != nil {
		return managed.ExternalCreation{}, nil
	}
	_, err = e.civoClient.CreateNewNetwork(cr.Spec.Name, cr.Spec.CIDRv4, cr.Spec.NameserversV4)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	cr.SetConditions(xpv1.Creating())

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CivoNetwork)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCivoNetwork)
	}

	// Retrieve the current state of the network
	currentNetwork, err := e.civoClient.GetNetwork(cr.Spec.Name)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	if currentNetwork == nil {
		return managed.ExternalUpdate{}, errors.New("network not found")
	}

	// Check if fields like NameserversV4, Name, or Label have changed
	if !stringSlicesEqual(currentNetwork.NameserversV4, cr.Spec.NameserversV4) ||
		currentNetwork.Label != cr.Spec.Name || currentNetwork.CIDR != cr.Spec.CIDRv4 {
		err := e.civoClient.UpdateNetwork(currentNetwork.ID, cr.Spec.Name, cr.Spec.CIDRv4, cr.Spec.NameserversV4)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateNetwork)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CivoNetwork)
	if !ok {
		return errors.New(errNotCivoNetwork)
	}
	err := e.civoClient.DeleteNetwork(cr.Status.AtProvider.ID)
	if err == nil {
		return nil
	}
	cr.SetConditions(xpv1.Deleting())

	return errors.Wrap(err, errDeleteNetwork)
}

// stringSlicesEqual returns true if two string slices are equal, false otherwise.
func stringSlicesEqual(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}
