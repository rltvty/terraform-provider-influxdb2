package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func testDataSourceOrganizationConfig(orgName string) string {
	return fmt.Sprintf(`
		resource "influxdb2_organization" "org" {
			name = "%s"
			description = "test org"
			status = "active"
		}
		data "influxdb2_organization" "by_name" {
			name = influxdb2_organization.org.name
		}
		data "influxdb2_organization" "by_id" {
			id = influxdb2_organization.org.id
		}
`, orgName)
}

func TestAccDataSourceOrganization(t *testing.T) {
	org := acctest.RandomWithPrefix("test-org")

	var provider *schema.Provider
	config := testConfig(testDataSourceOrganizationConfig(org))

	fmt.Println(config)

	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories(&provider),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.influxdb2_organization.by_id", "name", org),
					resource.TestCheckResourceAttr("data.influxdb2_organization.by_name", "name", org),
					resource.TestCheckResourceAttr("data.influxdb2_organization.by_id", "description", "test org"),
					resource.TestCheckResourceAttr("data.influxdb2_organization.by_name", "description", "test org"),
					resource.TestCheckResourceAttr("data.influxdb2_organization.by_id", "status", "active"),
					resource.TestCheckResourceAttr("data.influxdb2_organization.by_name", "status", "active"),
				),
			},
		},
	})
}
