package test

import (
	"fmt"
	"testing"

	"github.com/civo/civogo"
	. "github.com/onsi/gomega"
)

func TestReservedIPBasic(t *testing.T) {

	g := NewGomegaWithT(t)

	fmt.Println("Creating Reserved IP")
	ip, err := getOrCreateIP(e2eTest.civo)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		ip, err = e2eTest.civo.GetIP(ip.ID)
		return ip.IP
	}, "2m", "5s").ShouldNot(BeEmpty())
}

func TestIPAssignedToInstance(t *testing.T) {
	g := NewGomegaWithT(t)

	fmt.Println("Creating Reserved IP")
	ip, err := getOrCreateIP(e2eTest.civo)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		ip, err = e2eTest.civo.GetIP(ip.ID)
		return ip.IP
	}, "2m", "5s").ShouldNot(BeEmpty())

	fmt.Println("Creating Instance")
	instance, err := createInstance(e2eTest.civo)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Eventually(func() string {
		instance, err = e2eTest.civo.GetInstance(instance.ID)
		return instance.ReservedIP
	}, "2m", "5s").Should(Equal(ip.Name))

	g.Eventually(func() string {
		ip, err = e2eTest.civo.GetIP(ip.ID)
		return ip.AssignedTo.ID
	}, "2m", "5s").Should(Equal(instance.ID))
}

func getOrCreateIP(c *civogo.Client) (*civogo.IP, error) {
	ip, err := c.FindIP("e2e-test-ip")
	if err != nil && civogo.ZeroMatchesError.Is(err) {
		ip, err = c.NewIP(&civogo.CreateIPRequest{
			Name: "e2e-test-ip",
		})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return ip, err
}

func createInstance(c *civogo.Client) (*civogo.Instance, error) {
	instance, err := c.CreateInstance(&civogo.InstanceConfig{
		Hostname: "e2e-test-instance",
		Size:     "g3.xsmall",
	})
	if err != nil {
		return nil, err
	}
	return instance, err
}
