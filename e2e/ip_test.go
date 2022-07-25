package test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	civoinstance "github.com/crossplane-contrib/provider-civo/apis/civo/instance/v1alpha1"
	civoip "github.com/crossplane-contrib/provider-civo/apis/civo/ip/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReservedIPBasic(t *testing.T) {

	g := NewGomegaWithT(t)

	ip := &civoip.CivoIP{
		ObjectMeta: metav1.ObjectMeta{
			Name: "e2e-test-ip",
		},
	}

	fmt.Println("Creating Reserved IP")
	err := e2eTest.tenantClient.Create(context.TODO(), ip)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.Address
	}, "2m", "5s").ShouldNot(BeEmpty())
}

func TestIPAssignedToInstance(t *testing.T) {
	g := NewGomegaWithT(t)

	ip := &civoip.CivoIP{
		ObjectMeta: metav1.ObjectMeta{
			Name: "e2e-test-ip",
		},
	}

	fmt.Println("Creating Reserved IP")
	err := e2eTest.tenantClient.Create(context.TODO(), ip)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.Address
	}, "2m", "5s").ShouldNot(BeEmpty())

	instance := &civoinstance.CivoInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: "e2e-test-instance",
		},
		Spec: civoinstance.CivoInstanceSpec{
			InstanceConfig: civoinstance.CivoInstanceConfig{
				Hostname:  "e2e-test-instance",
				Size:      "g3.xsmall",
				DiskImage: "ubuntu-focal",
				Region:    "LON1",
			},
		},
	}

	fmt.Println("Creating Instance")
	err = e2eTest.tenantClient.Create(context.TODO(), instance)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(instance), instance)
		return instance.Spec.InstanceConfig.ReservedIP
	}, "2m", "5s").Should(Equal(ip.ObjectMeta.Name))

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.AssignedTo.ID
	}, "2m", "5s").Should(Equal(instance.Status.AtProvider.ID))
}
