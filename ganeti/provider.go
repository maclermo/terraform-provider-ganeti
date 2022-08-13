package ganeti

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GANETI_HOST", nil),
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  5080,
			},
			"api_version": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GANETI_USERNAME", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GANETI_PASSWORD", nil),
			},
			"ssl_verify": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"use_ssl": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"ganeti_instance": resourceInstance(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ganeti_instance": dataSourceInstance(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	host := d.Get("host").(string)
	port := d.Get("port").(int)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	apiVersion := d.Get("api_version").(int)
	useSsl := d.Get("use_ssl").(bool)
	sslVerify := d.Get("ssl_verify").(bool)

	c, err := NewClient(host, port, username, password, apiVersion, useSsl, sslVerify)

	if err != nil {
		return nil, err
	}

	return c, nil
}
