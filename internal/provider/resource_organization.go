package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/influxdata/influxdb-client-go/domain"
)

func resourceOrganization() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "The Organization resource allows you to configure a InfluxDB2 Organization.",

		CreateContext: resourceOrganizationCreate,
		ReadContext:   resourceOrganizationRead,
		UpdateContext: resourceOrganizationUpdate,
		DeleteContext: resourceOrganizationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: mergeSchemas(map[string]*schema.Schema{
			// Required Inputs
			"name": {
				Description: "Name of the Organization.",
				Type:        schema.TypeString,
				Required:    true,
			},
			// Optional Inputs
			"description": {
				Description: "The description of the Organization.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			// Computed outputs
			"id": {
				Description: "ID of the Organization.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		}, createdUpdatedSchema("Organization")),
	}
}

func resourceOrganizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*metaData).client
	orgsAPI := client.OrganizationsAPI()

	name := d.Get("name").(string)

	// Check for an existing Organization
	_, err := orgsAPI.FindOrganizationByName(ctx, name)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return diag.Errorf("unable to check for presence of an existing Organization (%s): %v", name, err)
		}
		log.Printf("[INFO] Organization (%s) not found, proceeding with create", name)
	} else {
		return diag.Errorf("unable to create Organization (%s) - an Organization with this name already exists; see resouce documentation for influxdb2_organization for instructions on how to add an already existing Organization to the state", name)
	}

	description := d.Get("description").(string)
	org := *&domain.Organization{
		Name:        name,
		Description: &description,
	}

	log.Printf("[INFO] Creating Organization (%s)", name)
	returnedOrg, err := orgsAPI.CreateOrganization(ctx, &org)
	if err != nil {
		return diag.Errorf("unable to create Organization (%s): %v", name, err)
	}

	if returnedOrg.Id == nil {
		return diag.Errorf("unable to create Organization (%s): <unknown error occurred>", name)
	}

	id := *returnedOrg.Id

	d.SetId(id)

	log.Printf("[INFO] Created Organization (%s) (%s)", name, id)

	// Get the updated Organization
	updatedOrg, err := orgsAPI.FindOrganizationByID(ctx, id)
	if err != nil {
		return diag.Errorf("unable to retrieve Organization (%s) (%s): %v", name, id, err)
	}

	if err := setOrganizationResourceData(d, updatedOrg); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*metaData).client
	orgsAPI := client.OrganizationsAPI()

	id := d.Id()

	log.Printf("[INFO] Reading Organization (%s)", id)

	org, err := orgsAPI.FindOrganizationByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[WARN] Organization (%s) not found, removing from state", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to retrieve Organization (%s): %v", id, err)
	}

	// Organization found, update resource data
	if err := setOrganizationResourceData(d, org); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOrganizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*metaData).client
	orgsAPI := client.OrganizationsAPI()

	id := d.Id()

	log.Printf("[INFO] Reading Organization (%s)", id)

	org, err := orgsAPI.FindOrganizationByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[WARN] Organization (%s) not found, removing from state", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to retrieve Organization (%s): %v", id, err)
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	org.Name = name
	org.Description = &description

	log.Printf("[INFO] Updating Organization (%s)", id)
	updatedOrg, err := orgsAPI.UpdateOrganization(ctx, org)
	if err != nil {
		return diag.Errorf("unable to update Organization (%s): %v", id, err)
	}

	log.Printf("[INFO] Updated Organization (%s)", id)

	if err := setOrganizationResourceData(d, updatedOrg); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOrganizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*metaData).client
	orgsAPI := client.OrganizationsAPI()

	id := d.Id()

	log.Printf("[INFO] Deleting Organization (%s)", id)

	err := orgsAPI.DeleteOrganizationWithID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[WARN] Organization (%s) not found, so no action was taken", id)
			return nil
		}
		return diag.Errorf("unable to delete Organization (%s): %v", id, err)
	}

	log.Printf("[INFO] Deleting (%s) deleted, removing from state", id)

	return nil
}

func setOrganizationResourceData(d *schema.ResourceData, org *domain.Organization) error {
	if err := d.Set("id", org.Id); err != nil {
		return err
	}
	if err := d.Set("name", org.Name); err != nil {
		return err
	}
	if err := d.Set("description", org.Description); err != nil {
		return err
	}
	if err := d.Set("created_at", org.CreatedAt.UTC().String()); err != nil {
		return err
	}
	if err := d.Set("updated_at", org.UpdatedAt.UTC().String()); err != nil {
		return err
	}
	if err := d.Set("created_timestamp", org.CreatedAt.Unix()); err != nil {
		return err
	}
	if err := d.Set("updated_timestamp", org.UpdatedAt.Unix()); err != nil {
		return err
	}
	return nil
}

// resourceOrganizationImport implements the logic necessary to import an un-tracked
// (by Terraform) Organization resource into Terraform state.
func resourceOrganizationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*metaData).client
	orgsAPI := client.OrganizationsAPI()

	id := d.Id()

	// Get the imported Organization
	importedOrg, err := orgsAPI.FindOrganizationByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to import Organization (%s) : %v", id, err)
	}

	if err := setOrganizationResourceData(d, importedOrg); err != nil {
		return nil, err
	}

	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}
