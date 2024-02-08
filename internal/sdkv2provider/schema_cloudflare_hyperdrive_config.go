package sdkv2provider

import (
	"fmt"

	"github.com/cloudflare/terraform-provider-cloudflare/internal/consts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var schemes = []string{"postgres", "postgresql"}

func resourceCloudflareHyperdriveConfigSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"account_id": {
			Description: consts.AccountIDSchemaDescription,
			Type:        schema.TypeString,
			Required:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the Hyperdrive configuration.",
		},
		"password": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The password of the Hyperdrive configuration.",
		},
		"origin": {
			Type:        schema.TypeSet,
			Required:    true,
			Description: "The origin details for the Hyperdrive configuration.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"database": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The name of your origin database.",
					},
					"host": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The host (hostname or IP) of your origin database.",
					},
					"port": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The port (default: 5432 for Postgres) of your origin database.",
						Default:    "5432",
					},
					"scheme": {
						Type:         schema.TypeString,
						Optional:     true,
						Description:  fmt.Sprintf("Specifies the URL scheme used to connect to your origin database. %s", renderAvailableDocumentationValuesStringSlice(schemes)),
						ValidateFunc: validation.StringInSlice(schemes, false),
						Default:      "postgres",
					},
					"user": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The user of your origin database.",
					},
				},
			},
		},
		"caching": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "The caching of the Hyperdrive.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"disabled": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "When set to true, disables the caching of SQL responses.",
						Default:     false,
					},
					"max_age": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "When present, specifies max duration for which items should persist in the cache.",
						Default:     60,
					},
					"stale_while_revalidate": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "When present, indicates the number of seconds cache may serve the response after it becomes stale.",
						Default:     15,
					},
				},
			},
		},
	}
}
