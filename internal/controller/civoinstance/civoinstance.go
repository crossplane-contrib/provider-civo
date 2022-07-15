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

package civoinstance

import (
	"context"

	"github.com/civo/civogo"
	v1alpha1provider "github.com/crossplane-contrib/provider-civo/apis/civo/provider/v1alpha1"
	"github.com/crossplane-contrib/provider-civo/pkg/civocli"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane-contrib/provider-civo/apis/civo/instance/v1alpha1"
	ipv1alpha1 "github.com/crossplane-contrib/provider-civo/apis/civo/ip/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/providerconfig"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errManagedUpdateFailed = "cannot update instance custom resource"
	errGenObservation      = "cannot generate observation"
	errNotCivoInstance     = "managed resource is not a CivoInstance"
	errCreateInstance      = "cannot create instance"
	errDeleteInstance      = "cannot delete instance"
	errGetSSHPubKeySecret  = "cannot get ssh public key secret %s"
	errUpdateInstance      = "cannot update instance"
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
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := providerconfig.ControllerName(v1alpha1.CivoInstancGroupKind)

	o := controller.Options{
		RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CivoInstancGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("civokubernetes", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.CivoInstance{}).
		Complete(r)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	instance, ok := mg.(*v1alpha1.CivoInstance)
	if !ok {
		return nil, errors.New(errNotCivoInstance)
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
	cr, ok := mg.(*v1alpha1.CivoInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCivoInstance)
	}
	civoInstance, err := e.civoClient.GetInstance(cr.Status.AtProvider.ID)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, err
	}
	if civoInstance == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider, err = civocli.GenerateObservation(civoInstance)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGenObservation)
	}

	switch civoInstance.Status {
	case civocli.StateActive:
		cr.SetConditions(xpv1.Available())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
			ConnectionDetails: managed.ConnectionDetails{
				xpv1.ResourceCredentialsSecretEndpointKey: []byte(civoInstance.PublicIP),
				xpv1.ResourceCredentialsSecretPortKey:     []byte("22"),
			},
		}, nil
	case civocli.StateBuilding:
		cr.SetConditions(xpv1.Creating())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}
	return managed.ExternalObservation{}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CivoInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCivoInstance)
	}
	cr.Status.SetConditions(xpv1.Creating())

	createInstance := cr.DeepCopy()

	var sshPubKey string

	if createInstance.Spec.InstanceConfig.SSHPubKeyRef != nil {
		s := &corev1.Secret{}
		n := types.NamespacedName{Namespace: createInstance.Spec.InstanceConfig.SSHPubKeyRef.Namespace, Name: createInstance.Spec.InstanceConfig.SSHPubKeyRef.Name}
		if err := e.kube.Get(ctx, n, s); err != nil {
			return managed.ExternalCreation{}, errors.Wrapf(err, errGetSSHPubKeySecret, n)
		}
		sshPubKey = string(s.Data[createInstance.Spec.InstanceConfig.SSHPubKeyRef.Key])
	}

	ip := &ipv1alpha1.CivoIP{}
	err := e.kube.Get(ctx, types.NamespacedName{Name: createInstance.Spec.InstanceConfig.ReservedIP}, ip)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	if ip.Status.AtProvider.ID != "" {
		ip, err := e.civoClient.GetIP(ip.Status.AtProvider.ID)
		if err != nil {
			return managed.ExternalCreation{}, err
		}
		err = e.civoClient.AssignIP(ip.ID, createInstance.Status.AtProvider.ID, "instance")
		if err != nil {
			return managed.ExternalCreation{}, err
		}
	} else {
		ip, err := findIPWithInstanceID(e.civoClient, createInstance.Status.AtProvider.ID)
		if err != nil {
			klog.Errorf("Unable to find IP with instance ID, error: %v", err)
			return managed.ExternalCreation{}, err
		}
		if ip != nil {
			err = e.civoClient.UnAssignIP(ip.ID)
			if err != nil {
				klog.Errorf("Unable to unassign IP, error: %v", err)
				return managed.ExternalCreation{}, err
			}
		}
	}

	instance, err := e.civoClient.CreateNewInstance(createInstance, sshPubKey, cr.Spec.InstanceConfig.DiskImage)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateInstance)
	}
	cr.Status.AtProvider.ID = instance.ID
	meta.SetExternalName(cr, instance.ID)
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errManagedUpdateFailed)
	}
	return managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(instance.PublicIP),
		xpv1.ResourceCredentialsSecretPortKey:     []byte("22"),
	}}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CivoInstance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCivoInstance)
	}

	err := e.civoClient.UpdateInstance(cr.Status.AtProvider.ID, cr)

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateInstance)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CivoInstance)
	if !ok {
		return errors.New(errNotCivoInstance)
	}
	cr.SetConditions(xpv1.Deleting())
	err := e.civoClient.DeleteInstance(cr.Status.AtProvider.ID)
	return errors.Wrap(err, errDeleteInstance)
}

func findIPWithInstanceID(civo *civocli.CivoClient, instanceID string) (*civogo.IP, error) {
	ips, err := civo.ListIPs()
	if err != nil {
		klog.Errorf("Unable to list IPs, error: %v", err)
		return nil, err
	}

	for _, ip := range ips.Items {
		if ip.AssignedTo.ID == instanceID {
			return &ip, nil
		}
	}
	return nil, nil
}
