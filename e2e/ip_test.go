package test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

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

	time.Sleep(15 * time.Second)

	newIP := &civoip.CivoIP{}
	err = e2eTest.tenantClient.Get(context.TODO(), types.NamespacedName{Name: "test-ip", Namespace: "crossplane-system"}, newIP)

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

	instance, err := getOrCreateInstance("e2e-test-instance", "test-ip")
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.AssignedTo.ID
	}, "2m", "5s").Should(BeEmpty())

	retry(30, 5*time.Second, func() error {
		e2eTest.tenantClient.Get(context.TODO(), client.ObjectKey{Name: "e2e-test-instance"}, instance)
		if instance.Status.AtProvider.ID == "" {
			return fmt.Errorf("instance id not updated yet")
		}
		return nil
	})
	g.Expect(instance.Status.AtProvider.ID).ShouldNot(BeEmpty())

	ip, err = assignIP(ip, instance.Status.AtProvider.ID)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.AssignedTo.ID
	}, "2m", "5s").Should(Equal(instance.Status.AtProvider.ID))
}

func TestIPUnassign(t *testing.T) {
	g := NewGomegaWithT(t)

	ip, err := getOrCreateIP("test-ip")
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.Address
	}, "2m", "5s").ShouldNot(BeEmpty())

	instance, err := getOrCreateInstance("e2e-test-instance", "test-ip")
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.AssignedTo.ID
	}, "2m", "5s").Should(BeEmpty())

	retry(30, 5*time.Second, func() error {
		e2eTest.tenantClient.Get(context.TODO(), client.ObjectKey{Name: "e2e-test-instance"}, instance)
		if instance.Status.AtProvider.ID == "" {
			return fmt.Errorf("instance id not updated yet")
		}
		return nil
	})
	g.Expect(instance.Status.AtProvider.ID).ShouldNot(BeEmpty())

	ip, err = assignIP(ip, instance.Status.AtProvider.ID)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.AssignedTo.ID
	}, "2m", "5s").Should(Equal(instance.Status.AtProvider.ID))

	instance, err = unassignIP(instance)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(instance), instance)
		return instance.Spec.InstanceConfig.ReservedIP
	}, "2m", "5s").Should(Equal(""))
}

func TestDeleteIP(t *testing.T) {
	g := NewGomegaWithT(t)

	ip, err := getOrCreateIP("test-ip")
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		err = e2eTest.tenantClient.Get(context.TODO(), client.ObjectKeyFromObject(ip), ip)
		return ip.Status.AtProvider.Address
	}, "2m", "5s").ShouldNot(BeEmpty())

	err = deleteIP(ip.Name)
	g.Expect(err).ShouldNot(HaveOccurred())
}

func deleteIP(name string) error {
	ip := &civoip.CivoIP{}
	err := e2eTest.tenantClient.Get(context.TODO(), client.ObjectKey{Name: name}, ip)
	if errors.IsNotFound(err) {
		return nil
	}
	err = e2eTest.tenantClient.Delete(context.TODO(), ip)
	return err
}

func getOrCreateIP(name string) (*civoip.CivoIP, error) {
	ip := &civoip.CivoIP{}
	err := e2eTest.tenantClient.Get(context.TODO(), client.ObjectKey{Name: name}, ip)
	if errors.IsNotFound(err) {
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
	return ip, nil
}

func assignIP(ip *civoip.CivoIP, instanceID string) (*civoip.CivoIP, error) {

	ip.Status.AtProvider.AssignedTo.ID = instanceID
	err := e2eTest.tenantClient.Status().Update(context.TODO(), ip)

	if err != nil {
		return nil, err
	}
	return ip, nil
}

func unassignIP(instance *civoinstance.CivoInstance) (*civoinstance.CivoInstance, error) {

	instance.Spec.InstanceConfig.ReservedIP = ""
	err := e2eTest.tenantClient.Status().Update(context.TODO(), instance)

	if err != nil {
		return nil, err
	}
	return instance, nil
}

func getOrCreateInstance(name string, reservedIPName string) (*civoinstance.CivoInstance, error) {
	instance := &civoinstance.CivoInstance{}
	err := e2eTest.tenantClient.Get(context.TODO(), client.ObjectKey{Name: name}, instance)

	if errors.IsNotFound(err) {
		fmt.Println("Creating Instance")
		instance = &civoinstance.CivoInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:         name,
				GenerateName: name,
				Namespace:    "crossplane-system",
			},
			Spec: civoinstance.CivoInstanceSpec{
				InstanceConfig: civoinstance.CivoInstanceConfig{
					ReservedIP: reservedIPName,
					Size:       "g3.xsmall",
					DiskImage:  "ubuntu-focal",
					Region:     "LON1",
					Hostname:   randomHostname(),
				},
				ResourceSpec: xpv1.ResourceSpec{
					ProviderConfigReference: &xpv1.Reference{
						Name: "civo-provider",
					},
				},
			},
		}
		err = e2eTest.tenantClient.Create(context.TODO(), instance)
		return instance, err
	} else if err != nil {
		return nil, err
	}
	return instance, nil
}

func randomHostname() string {
	var chars = []rune("0123456789")
	rand.Seed(time.Now().UnixNano())

	s := make([]rune, 4)
	for i := range s {
		s[i] = chars[rand.Intn(len(chars))]
	}
	return "hostname-" + string(s)
}
