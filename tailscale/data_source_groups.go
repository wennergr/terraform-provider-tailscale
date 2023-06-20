package tailscale

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/tailscale/tailscale-client-go/tailscale"
)

func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "The group data source describes a single group in the policy",
		ReadContext: readWithWaitFor(dataSourceGroupRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the group",
			},
			"members": {
				Type:        schema.TypeList,
				Description: "members of the group",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"wait_for": {
				Type:        schema.TypeString,
				Description: "If specified, the provider will make multiple attempts to obtain the data source until the wait_for duration is reached. Retries are made every second so this value should be greater than 1s",
				Optional:    true,
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					waitFor, err := time.ParseDuration(i.(string))
					switch {
					case err != nil:
						return diagnosticsErrorWithPath(err, "failed to parse wait_for", path)
					case waitFor <= time.Second:
						return diagnosticsErrorWithPath(nil, "wait_for must be greater than 1 second", path)
					default:
						return nil
					}
				},
			},
		},
	}
}

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*tailscale.Client)
	name := d.Get("name").(string)

	acl, err := client.ACL(ctx)
	if err != nil {
		return diagnosticsError(err, "Failed to fetch groups")
	}

	members, ok := acl.Groups[name]
	if !ok {
		return diagnosticsError(nil, "Unable to find group %s", name)
	}

	d.SetId(fmt.Sprintf("acl.groups.%s", name))
	if err = d.Set("members", members); err != nil {
		return diagnosticsError(err, "Failed to set group members")
	}

	return nil
}
