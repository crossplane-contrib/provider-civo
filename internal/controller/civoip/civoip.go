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

package civoip

import (
	"context"
	"strings"

	"github.com/civo/civogo"
	"github.com/crossplane-contrib/provider-civo/apis/civo/ip/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/providerconfig"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	v1alpha1provider "github.com/crossplane-contrib/provider-civo/apis/civo/provider/v1alpha1"
)

const (
	errNotCivoIP = "not a CivoIP resource"
	errRenameIP  = "cannot rename IP"
	errDeleteIP  = "cannot delete IP"
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
	name := providerconfig.ControllerName(v1alpha1.CivoIPGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CivoIPGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("civoip", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.CivoIP{}).
		Complete(r)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	ip, ok := mg.(*v1alpha1.CivoIP)
	if !ok {
		return nil, errors.New(errNotCivoIP)
	}

	providerConfig := &v1alpha1provider.ProviderConfig{}

	err := c.client.Get(ctx, types.NamespacedName{
		Name: ip.Spec.ProviderConfigReference.Name}, providerConfig)

	if err != nil {
		return nil, err
	}

	s := &corev1.Secret{}
	if err := c.client.Get(ctx, types.NamespacedName{Name: providerConfig.Spec.Credentials.SecretRef.Name,
		Namespace: providerConfig.Spec.Credentials.SecretRef.Namespace}, s); err != nil {
		return nil, errors.New("could not find secret")
	}

	apiKey := string(s.Data["credentials"])
	region := providerConfig.Spec.Region
	if apiKey == "" {
		return nil, errors.New("newCivoClient: apiKey is nil")
	}

	apiKey = strings.TrimSuffix(apiKey, "\n")

	if region == "" {
		return nil, errors.New("newCivoClient: region is nil")
	}
	client, err := civogo.NewClient(apiKey, region)
	if err != nil {
		return nil, err
	}
	return &external{
		kube:         c.client,
		civoGoClient: client,
	}, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CivoIP)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCivoIP)
	}

	civoIP, err := findIPExact(e.civoGoClient, cr.Name)
	if err != nil {
		if strings.Contains(err.Error(), "DatabaseIPNotFoundError") {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{ResourceExists: false}, err
	}
	// if civoIP == nil {
	// 	return managed.ExternalObservation{ResourceExists: false}, nil
	// }
	if civoIP.IP != "" {
		cr.Status.AtProvider.ID = civoIP.IP
		cr.Status.AtProvider.AssignedTo.ID = civoIP.AssignedTo.ID
		cr.Status.AtProvider.AssignedTo.Name = civoIP.AssignedTo.Name
		cr.Status.AtProvider.AssignedTo.Type = civoIP.AssignedTo.Type
		cr.SetConditions(xpv1.Available())
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	} else {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CivoIP)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCivoIP)
	}

	civoIP, err := findIPExact(e.civoGoClient, cr.Name)
	if err != nil {
		if !strings.Contains(err.Error(), "DatabaseIPNotFoundError") {
			return managed.ExternalCreation{}, err
		}
	}
	if civoIP != nil {
		return managed.ExternalCreation{}, nil
	}

	providerConfig := &v1alpha1provider.ProviderConfig{}

	err = e.kube.Get(ctx, types.NamespacedName{
		Name: cr.Spec.ProviderConfigReference.Name}, providerConfig)

	if err != nil {
		return managed.ExternalCreation{}, err
	}

	_, err = e.civoGoClient.NewIP(&civogo.CreateIPRequest{
		Name:   cr.Name,
		Region: providerConfig.Spec.Region,
	})
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
	cr, ok := mg.(*v1alpha1.CivoIP)
	if !ok {
		return errors.New(errNotCivoIP)
	}
	cr.SetConditions(xpv1.Deleting())
	civoIP, err := findIPExact(e.civoGoClient, cr.Name)
	if err != nil {
		if strings.Contains(err.Error(), "DatabaseIPNotFoundError") {
			return nil
		}
		return err
	}
	_, err = e.civoGoClient.DeleteIP(civoIP.ID)
	return err
}

func findIPExact(client *civogo.Client, name string) (*civogo.IP, error) {
	ipList, err := client.ListIPs()
	if err != nil {
		return nil, err
	}
	for _, ip := range ipList.Items {
		if ip.Name == name {
			return &ip, nil
		}
	}
	return nil, errors.New("DatabaseIPNotFoundError")
}
