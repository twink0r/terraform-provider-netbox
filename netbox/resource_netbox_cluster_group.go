package netbox

import (
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/virtualization"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxClusterGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxClusterGroupCreate,
		Read:   resourceNetboxClusterGroupRead,
		Update: resourceNetboxClusterGroupUpdate,
		Delete: resourceNetboxClusterGroupDelete,

		Description: `From the [official documentation](https://docs.netbox.dev/en/stable/core-functionality/virtualization/#cluster-groups):

> Cluster groups may be created for the purpose of organizing clusters. The arrangement of clusters into groups is optional.`,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 30),
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceNetboxClusterGroupCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	data := models.ClusterGroup{}

	name := d.Get("name").(string)
	data.Name = &name

	slugValue, slugOk := d.GetOk("slug")
	var slug string
	// Default slug to name attribute if not given
	if !slugOk {
		slug = name
	} else {
		slug = slugValue.(string)
	}
	data.Slug = &slug

	if description, ok := d.GetOk("description"); ok {
		data.Description = description.(string)
	}

	params := virtualization.NewVirtualizationClusterGroupsCreateParams().WithData(&data)

	res, err := api.Virtualization.VirtualizationClusterGroupsCreate(params, nil)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(res.GetPayload().ID, 10))

	return resourceNetboxClusterGroupRead(d, m)
}

func resourceNetboxClusterGroupRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := virtualization.NewVirtualizationClusterGroupsReadParams().WithID(id)

	res, err := api.Virtualization.VirtualizationClusterGroupsRead(params, nil)
	if err != nil {
		errorcode := err.(*virtualization.VirtualizationClusterGroupsReadDefault).Code()
		if errorcode == 404 {
			// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", res.GetPayload().Name)
	d.Set("slug", res.GetPayload().Slug)
	d.Set("description", res.GetPayload().Description)
	return nil
}

func resourceNetboxClusterGroupUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	data := models.ClusterGroup{}

	name := d.Get("name").(string)
	data.Name = &name

	slugValue, slugOk := d.GetOk("slug")
	var slug string
	// Default slug to name if not given
	if !slugOk {
		slug = name
	} else {
		slug = slugValue.(string)
	}
	data.Slug = &slug

	if d.HasChange("description") {
		// description omits empty values so set to ' '
		if description := d.Get("description"); description.(string) == "" {
			data.Description = " "
		} else {
			data.Description = description.(string)
		}
	}

	params := virtualization.NewVirtualizationClusterGroupsPartialUpdateParams().WithID(id).WithData(&data)

	_, err := api.Virtualization.VirtualizationClusterGroupsPartialUpdate(params, nil)
	if err != nil {
		return err
	}

	return resourceNetboxClusterGroupRead(d, m)
}

func resourceNetboxClusterGroupDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := virtualization.NewVirtualizationClusterGroupsDeleteParams().WithID(id)

	_, err := api.Virtualization.VirtualizationClusterGroupsDelete(params, nil)
	if err != nil {
		return err
	}
	return nil
}
