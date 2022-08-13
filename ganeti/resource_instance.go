package ganeti

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,

		Schema: map[string]*schema.Schema{
			"admin_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Instance's administrative state",
			},
			"disk": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "Disks to create (needs at least 1)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Disk size, eg: 20G",
						},
					},
				},
			},
			"disk_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the disk template to use for the storage backend",
			},
			"group_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the group name to use for clustering",
			},
			"hypervisor": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Type of hypervisor to spawn instances on (use kvm if unsure)",
			},
			"memory": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Memory of the instance",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the instance",
			},
			"network": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "Networks to create (needs at least 1)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"link": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Network link to bind to the hypervisor",
						},
					},
				},
			},
			"node": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Node to spawn the instance on",
			},
			"os_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Bootstrap images and automation scripts",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Instance's status",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Instance's UUID",
			},
			"vcpus": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Number of virtual cpus to allocate to the instance",
			},
		},
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	var instanceDisks []Disk

	diskInterface := d.Get("disk").([]interface{})

	for _, disk := range diskInterface {
		var instanceDisk Disk

		i := disk.(map[string]interface{})

		instanceDisk.Size = i["size"].(string)

		instanceDisks = append(instanceDisks, instanceDisk)
	}

	var instanceNetworks []NIC

	networkInterface := d.Get("network").([]interface{})

	for _, network := range networkInterface {
		var instanceNetwork NIC

		i := network.(map[string]interface{})

		instanceNetwork.Link = i["link"].(string)

		instanceNetworks = append(instanceNetworks, instanceNetwork)
	}

	backendParams := BackendParams{
		Memory: d.Get("memory").(string),
		VCPUs:  d.Get("vcpus").(int),
	}

	instance := Instance{
		Version:       1,
		BackendParams: backendParams,
		Disks:         instanceDisks,
		DiskTemplate:  d.Get("disk_template").(string),
		GroupName:     d.Get("group_name").(string),
		Hypervisor:    d.Get("hypervisor").(string),
		Name:          d.Get("name").(string),
		Mode:          "create",
		NICs:          instanceNetworks,
		OSType:        d.Get("os_type").(string),
		Node:          d.Get("node").(string),
	}

	result, err := c.CreateInstance(instance)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(result.Name)

	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	var diags diag.Diagnostics

	id := d.Get("name").(string)

	result, err := c.ReadInstance(id)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("admin_state", result.AdminState)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("disk_template", result.DiskTemplate)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("memory", transformSize(result.BackendParams.Memory))
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("name", result.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("node", result.Node)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("os_type", result.OSType)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("status", result.Status)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("uuid", result.UUID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("vcpus", result.BackendParams.VCPUs)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(result.Name)

	return diags
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.FromErr(fmt.Errorf("Not implemented"))
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	var diags diag.Diagnostics

	id := d.Get("name").(string)

	err := c.DeleteInstance(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func transformSize(n int) string {
	var size string

	if val := n / 1024 / 1024 / 1024; val > 0 {
		size = fmt.Sprintf("%dP", val)
	} else if val := n / 1024 / 1024; val > 0 {
		size = fmt.Sprintf("%dT", val)
	} else if val := n / 1024; val > 0 {
		size = fmt.Sprintf("%dG", val)
	} else {
		size = fmt.Sprintf("%dM", val)
	}

	return size
}
