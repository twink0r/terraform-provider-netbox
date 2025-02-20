package netbox

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/ipam"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccNetboxIPAddressFullDependencies(testName string) string {
	return fmt.Sprintf(`
resource "netbox_tag" "test" {
  name = "%[1]s"
}

resource "netbox_tenant" "test" {
  name = "%[1]s"
}

resource "netbox_vrf" "test" {
  name = "%[1]s"
}

resource "netbox_cluster_type" "test" {
  name = "%[1]s"
}

resource "netbox_cluster" "test" {
  name = "%[1]s"
  cluster_type_id = netbox_cluster_type.test.id
}

resource "netbox_virtual_machine" "test" {
  name = "%[1]s"
  cluster_id = netbox_cluster.test.id
}

resource "netbox_interface" "test" {
  name = "%[1]s"
  virtual_machine_id = netbox_virtual_machine.test.id
}
`, testName)
}

func TestAccNetboxIPAddress_basic(t *testing.T) {

	testIP := "1.1.1.1/32"
	testSlug := "ipaddress"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxIPAddressFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_ip_address" "test" {
  ip_address = "%s"
  interface_id = netbox_interface.test.id
  status = "active"
  tags = [netbox_tag.test.name]
}`, testIP),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", testIP),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "tags.0", testName),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "tenant_id", "0"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "vrf_id", "0"),
				),
			},
			{
				Config: testAccNetboxIPAddressFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_ip_address" "test" {
  ip_address = "%s"
  interface_id = netbox_interface.test.id
  status = "reserved"
  tenant_id = netbox_tenant.test.id
  vrf_id = netbox_vrf.test.id
  tags = [netbox_tag.test.name]
  description = "description for %[1]s"
}`, testIP),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", testIP),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "reserved"),
					resource.TestCheckResourceAttrPair("netbox_ip_address.test", "tenant_id", "netbox_tenant.test", "id"),
					resource.TestCheckResourceAttrPair("netbox_ip_address.test", "vrf_id", "netbox_vrf.test", "id"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "description", fmt.Sprintf("description for %[1]s", testIP)),
				),
			},
			{
				Config: testAccNetboxIPAddressFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_ip_address" "test" {
  ip_address = "%s"
  interface_id = netbox_interface.test.id
  status = "dhcp"
  tags = [netbox_tag.test.name]
}`, testIP),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", testIP),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "dhcp"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "tenant_id", "0"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "vrf_id", "0"),
				),
			},
			{
				Config: testAccNetboxIPAddressFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_ip_address" "test" {
  ip_address = "%s"
  interface_id = netbox_interface.test.id
  status = "provoke_error"
  tags = [netbox_tag.test.name]
}`, testIP),
				ExpectError: regexp.MustCompile("expected status to be one of .*"),
			},
			{
				Config: testAccNetboxIPAddressFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_ip_address" "test" {
  ip_address = "%s"
  interface_id = netbox_interface.test.id
  status = "deprecated"
  tags = [netbox_tag.test.name]
}`, testIP),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", testIP),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "deprecated"),
				),
			},
			{
				Config: testAccNetboxIPAddressFullDependencies(testName) + fmt.Sprintf(`
resource "netbox_ip_address" "test" {
  ip_address = "%s"
  interface_id = netbox_interface.test.id
  status = "active"
  dns_name = "mytest.example.com"
  tags = [netbox_tag.test.name]
}`, testIP),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_ip_address.test", "ip_address", testIP),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("netbox_ip_address.test", "dns_name", "mytest.example.com"),
				),
			},
			{
				ResourceName:      "netbox_ip_address.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_ip_address", &resource.Sweeper{
		Name:         "netbox_ip_address",
		Dependencies: []string{},
		F: func(region string) error {
			m, err := sharedClientForRegion(region)
			if err != nil {
				return fmt.Errorf("Error getting client: %s", err)
			}
			api := m.(*client.NetBoxAPI)
			params := ipam.NewIpamIPAddressesListParams()
			res, err := api.Ipam.IpamIPAddressesList(params, nil)
			if err != nil {
				return err
			}
			for _, ipAddress := range res.GetPayload().Results {
				if len(ipAddress.Tags) > 0 && (ipAddress.Tags[0] == &models.NestedTag{Name: strToPtr("acctest"), Slug: strToPtr("acctest")}) {
					deleteParams := ipam.NewIpamIPAddressesDeleteParams().WithID(ipAddress.ID)
					_, err := api.Ipam.IpamIPAddressesDelete(deleteParams, nil)
					if err != nil {
						return err
					}
					log.Print("[DEBUG] Deleted an ip address")
				}
			}
			return nil
		},
	})
}
