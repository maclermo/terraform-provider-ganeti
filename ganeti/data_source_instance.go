package ganeti

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInstanceRead,
		Schema:      nil,
	}
}

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}
