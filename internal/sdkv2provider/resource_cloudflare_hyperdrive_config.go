package sdkv2provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/cloudflare/terraform-provider-cloudflare/internal/consts"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceCloudflareHyperdriveConfig() *schema.Resource {
	return &schema.Resource{
		Schema:        resourceCloudflareHyperdriveConfigSchema(),
		CreateContext: resourceCloudflareHyperdriveConfigCreate,
		ReadContext:   resourceCloudflareHyperdriveConfigRead,
		UpdateContext: resourceCloudflareHyperdriveConfigUpdate,
		DeleteContext: resourceCloudflareHyperdriveConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceCloudflareHyperdriveConfigImport,
		},
		Description: "Provides the ability to manage Cloudflare Hyperdrive configurations.",
	}
}

func resourceCloudflareHyperdriveConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflare.API)
	accountID := d.Get(consts.AccountIDSchemaKey).(string)

	var newHyperdriveConfig cloudflare.CreateHyperdriveConfigParams

	if name, ok := d.GetOk("name"); ok {
		newHyperdriveConfig.Name = name.(string)
	}

	if password, ok := d.GetOk("password"); ok {
		newHyperdriveConfig.Password = password.(string)
	}

	if origin, ok := d.GetOk("origin"); ok {
		newHyperdriveConfig.Origin = origin.(cloudflare.HyperdriveConfigOrigin)
	}

	if caching, ok := d.GetOk("caching"); ok {
		newHyperdriveConfig.Caching = caching.(cloudflare.HyperdriveConfigCaching)
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating Cloudflare Hyperdrive Config from struct: %+v", newHyperdriveConfig))

	var r cloudflare.HyperdriveConfig
	var err error

	r, err = client.CreateHyperdriveConfig(ctx, cloudflare.AccountIdentifier(accountID), newHyperdriveConfig)

	if err != nil {
		return diag.FromErr(errors.Wrap(err, "error creating hyperdrive config"))
	}

	d.SetId(r.ID)

	tflog.Info(ctx, fmt.Sprintf("Cloudflare Hyperdrive Config ID: %s", d.Id()))

	return resourceCloudflareHyperdriveConfigRead(ctx, d, meta)
}

func resourceCloudflareHyperdriveConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflare.API)
	hyperdriveConfigID := d.Id()
	accountID := d.Get(consts.AccountIDSchemaKey).(string)

	resp, err := client.GetHyperdriveConfig(ctx, cloudflare.AccountIdentifier(accountID), hyperdriveConfigID)
	if err != nil {
		var notFoundError *cloudflare.NotFoundError
		if errors.As(err, &notFoundError) {
			tflog.Info(ctx, fmt.Sprintf("Hyperdrive Config %s no longer exists", d.Id()))
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error finding Hyperdrive Config %q: %w", d.Id(), err))
	}

	d.Set("name", resp.Name)

	if err := d.Set("origin", resp.Origin); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set origin: %w", err))
	}

	if err := d.Set("caching", resp.Caching); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set caching: %w", err))
	}

	return nil
}

func resourceCloudflareHyperdriveConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflare.API)
	accountID := d.Get(consts.AccountIDSchemaKey).(string)

	var newHyperdriveConfig cloudflare.UpdateHyperdriveConfigParams
	newHyperdriveConfig.HyperdriveID = d.Id()

	if name, ok := d.GetOk("name"); ok {
		newHyperdriveConfig.Name = name.(string)
	}

	if password, ok := d.GetOk("password"); ok {
		newHyperdriveConfig.Password = password.(string)
	}

	if origin, ok := d.GetOk("origin"); ok {
		newHyperdriveConfig.Origin = origin.(cloudflare.HyperdriveConfigOrigin)
	}

	if caching, ok := d.GetOk("caching"); ok {
		newHyperdriveConfig.Caching = caching.(cloudflare.HyperdriveConfigCaching)
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating Cloudflare Hyperdrive Config from struct: %+v", newHyperdriveConfig))

	r, err := client.UpdateHyperdriveConfig(ctx, cloudflare.AccountIdentifier(accountID), newHyperdriveConfig)

	if err != nil {
		return diag.FromErr(errors.Wrap(err, "error updating hyperdrive config"))
	}

	if r.ID == "" {
		return diag.FromErr(fmt.Errorf("failed to find id in Update response; resource was empty"))
	}

	return resourceCloudflareHyperdriveConfigRead(ctx, d, meta)
}

func resourceCloudflareHyperdriveConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*cloudflare.API)
	accountID := d.Get(consts.AccountIDSchemaKey).(string)

	tflog.Info(ctx, fmt.Sprintf("Deleting Cloudflare Hyperdrive Config with id: %+v", d.Id()))

	err := client.DeleteHyperdriveConfig(ctx, cloudflare.AccountIdentifier(accountID), d.Id())
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "error deleting hyperdrive config"))
	}

	d.SetId("")
	return nil
}

func resourceCloudflareHyperdriveConfigImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)
	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"accountID/hyperdriveConfigID\"", d.Id())
	}

	accountID, hyperdriveConfigID := attributes[0], attributes[1]
	tflog.Debug(ctx, fmt.Sprintf("Importing Cloudflare Hyperdrive Config id %s for account %s", hyperdriveConfigID, accountID))

	d.Set(consts.AccountIDSchemaKey, accountID)
	d.SetId(hyperdriveConfigID)

	resourceCloudflareHyperdriveConfigRead(ctx, d, meta)
	return []*schema.ResourceData{d}, nil
}
