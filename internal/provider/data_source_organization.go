package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/influxdata/influxdb-client-go/domain"
)

func dataSourceOrganization() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Lookup an Organization in InfluxDB2.",

		ReadContext: dataSourceOrganizationRead,

		Schema: mergeSchemas(map[string]*schema.Schema{
			// Optional inputs
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the Organization.",
			},
			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID of the Organization.",
			},
			// Computed outputs
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the Organization.",
			},
		}, createdUpdatedSchema("Organization")),
	}
}

func dataSourceOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*metaData).client
	orgAPI := client.OrganizationsAPI()

	// Warning or errors can be collected in a slice type
	var (
		diags diag.Diagnostics
		org   *domain.Organization
		err   error
	)

	if v, ok := d.GetOk("name"); ok {
		orgName := v.(string)
		if org, err = orgAPI.FindOrganizationByName(ctx, orgName); err != nil {
			diags = append(diags, diag.FromErr(err)...)
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Can't find Organization with name: %s", orgName),
			})
			return diags
		}
	} else if v, ok := d.GetOk("id"); ok {
		orgID := v.(string)
		if org, err = orgAPI.FindOrganizationByID(ctx, orgID); err != nil {
			diags = append(diags, diag.FromErr(err)...)
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Can't find Organization with id: %s", orgID),
			})
			return diags
		}
	}

	id := org.Id
	if id == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Organization not found",
		})
		return diags
	}

	d.SetId(*id)
	d.Set("id", *id)
	d.Set("name", org.Name)
	if org.Description != nil {
		d.Set("description", *org.Description)
	}

	return diags
}
