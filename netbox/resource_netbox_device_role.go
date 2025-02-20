package netbox

import (
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/dcim"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDeviceRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDeviceRoleCreate,
		Read:   resourceNetboxDeviceRoleRead,
		Update: resourceNetboxDeviceRoleUpdate,
		Delete: resourceNetboxDeviceRoleDelete,

		Description: `From the [official documentation](https://docs.netbox.dev/en/stable/core-functionality/devices/#device-roles):

> Devices can be organized by functional roles, which are fully customizable by the user. For example, you might create roles for core switches, distribution switches, and access switches within your network.`,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vm_role": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"color_hex": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceNetboxDeviceRoleCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	name := d.Get("name").(string)
	slugValue, slugOk := d.GetOk("slug")
	var slug string

	// Default slug to name if not given
	if !slugOk {
		slug = name
	} else {
		slug = slugValue.(string)
	}

	color := d.Get("color_hex").(string)
	vmRole := d.Get("vm_role").(bool)

	params := dcim.NewDcimDeviceRolesCreateParams().WithData(
		&models.DeviceRole{
			Name:   &name,
			Slug:   &slug,
			Color:  color,
			VMRole: vmRole,
		},
	)

	res, err := api.Dcim.DcimDeviceRolesCreate(params, nil)
	if err != nil {
		//return errors.New(getTextFromError(err))
		return err
	}

	d.SetId(strconv.FormatInt(res.GetPayload().ID, 10))

	return resourceNetboxDeviceRoleRead(d, m)
}

func resourceNetboxDeviceRoleRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := dcim.NewDcimDeviceRolesReadParams().WithID(id)

	res, err := api.Dcim.DcimDeviceRolesRead(params, nil)
	if err != nil {
		errorcode := err.(*dcim.DcimDeviceRolesReadDefault).Code()
		if errorcode == 404 {
			// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", res.GetPayload().Name)
	d.Set("slug", res.GetPayload().Slug)
	d.Set("vm_role", res.GetPayload().VMRole)
	d.Set("color_hex", res.GetPayload().Color)
	return nil
}

func resourceNetboxDeviceRoleUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	data := models.DeviceRole{}

	name := d.Get("name").(string)
	color := d.Get("color_hex").(string)
	vmRole := d.Get("vm_role").(bool)

	slugValue, slugOk := d.GetOk("slug")
	var slug string

	// Default slug to name if not given
	if !slugOk {
		slug = name
	} else {
		slug = slugValue.(string)
	}

	data.Slug = &slug
	data.Name = &name
	data.VMRole = vmRole
	data.Color = color

	params := dcim.NewDcimDeviceRolesPartialUpdateParams().WithID(id).WithData(&data)

	_, err := api.Dcim.DcimDeviceRolesPartialUpdate(params, nil)
	if err != nil {
		return err
	}

	return resourceNetboxDeviceRoleRead(d, m)
}

func resourceNetboxDeviceRoleDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := dcim.NewDcimDeviceRolesDeleteParams().WithID(id)

	_, err := api.Dcim.DcimDeviceRolesDelete(params, nil)
	if err != nil {
		return err
	}
	return nil
}
