package test

import (
	"context"
	"fmt"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	civoinstance "github.com/crossplane-contrib/provider-civo/apis/civo/instance/v1alpha1"
	civoip "github.com/crossplane-contrib/provider-civo/apis/civo/ip/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReservedIPBasic(t *testing.T) {
	g := NewGomegaWithT(t)

	_, err := getOrCreateIP("test-ip")
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		newIP := &civoip.CivoIP{}
		err = e2eTest.tenantClient.Get(context.TODO(), types.NamespacedName{Name: "test-ip", Namespace: "crossplane-system"}, newIP)
		return newIP.Status.AtProvider.ID
	}, "2m", "5s").ShouldNot(BeEmpty())

	g.Eventually(func() string {
		newIP := &civoip.CivoIP{}
		err = e2eTest.tenantClient.Get(context.TODO(), types.NamespacedName{Name: "test-ip", Namespace: "crossplane-system"}, newIP)
		return newIP.Status.AtProvider.Address
	}, "2m", "5s").ShouldNot(BeEmpty())
}

func TestIPAssignedToInstance(t *testing.T) {
	g := NewGomegaWithT(t)

	ip, err := getOrCreateIP("test-ip")
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.Address
	}, "2m", "5s").ShouldNot(BeEmpty())

	instance := &civoinstance.CivoInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-test-instance",
			Namespace: "crossplane-system",
		},
		Spec: civoinstance.CivoInstanceSpec{
			InstanceConfig: civoinstance.CivoInstanceConfig{
				ReservedIP: "test-ip",
				Size:       "g3.xsmall",
				DiskImage:  "ubuntu-focal",
				Region:     "LON1",
			},
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{
					Name: "civo-provider",
				},
			},
		},
	}

	fmt.Println("Creating Instance")
	err = e2eTest.tenantClient.Create(context.TODO(), instance)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.AssignedTo.ID
	}, "2m", "5s").Should(Equal(instance.Status.AtProvider.ID))
}

func getOrCreateIP(name string) (*civoip.CivoIP, error) {
	ip := &civoip.CivoIP{}
	err := e2eTest.tenantClient.Get(context.TODO(), client.ObjectKey{Name: name}, ip)
	if err != nil && errors.IsNotFound(err) {
		ip = &civoip.CivoIP{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "crossplane-system",
			},
			Spec: civoip.CivoIPSpec{
				ResourceSpec: xpv1.ResourceSpec{
					ProviderConfigReference: &xpv1.Reference{
						Name: "civo-provider",
					},
				},
			},
		}
		fmt.Println("Creating Reserved IP")
		err = e2eTest.tenantClient.Create(context.TODO(), ip)
		return ip, err
	} else if err != nil {
		return nil, err
	}
	return nil, err
}
