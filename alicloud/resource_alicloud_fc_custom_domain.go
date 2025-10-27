package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudFCCustomDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFCCustomDomainCreate,
		Read:   resourceAliCloudFCCustomDomainRead,
		Update: resourceAliCloudFCCustomDomainUpdate,
		Delete: resourceAliCloudFCCustomDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_info": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								equal, _ := compareJsonTemplateAreEquivalent(old, new)
								return equal
							},
						},
						"auth_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: StringInSlice([]string{"anonymous", "function", "jwt"}, false),
						},
					},
				},
			},
			"cert_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"cert_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"certificate": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_domain_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: StringInSlice([]string{"HTTP", "HTTPS", "HTTP,HTTPS"}, false),
			},
			"route_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"routes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"path": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"function_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"qualifier": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"rewrite_config": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"wildcard_rules": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"replacement": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"match": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
												"regex_rules": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"replacement": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"match": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
												"equal_rules": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"replacement": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"match": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"methods": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"subdomain_count": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tls_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: StringInSlice([]string{"TLSv1.3", "TLSv1.2", "TLSv1.1", "TLSv1.0"}, false),
						},
						"max_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: StringInSlice([]string{"TLSv1.3", "TLSv1.2", "TLSv1.1", "TLSv1.0"}, false),
						},
						"cipher_suites": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"waf_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_waf": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudFCCustomDomainCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	log.Printf("[DEBUG] Creating FC Custom Domain")

	// Build custom domain from schema
	domain := fcService.BuildCreateCustomDomainInputFromSchema(d)

	// Create custom domain using service layer
	var result *aliyunFCAPI.CustomDomain
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		result, err = fcService.CreateFCCustomDomain(domain)
		if err != nil {
			if NeedRetry(err) {
				log.Printf("[WARN] FC Custom Domain creation failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Custom Domain creation failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_fc_custom_domain", "CreateCustomDomain", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	if result != nil && result.DomainName != nil {
		d.SetId(*result.DomainName)
		log.Printf("[DEBUG] FC Custom Domain created successfully: %s", *result.DomainName)
	} else {
		return WrapErrorf(fmt.Errorf("failed to get domain name from create response"), DefaultErrorMsg, "alicloud_fc_custom_domain", "CreateCustomDomain", AlibabaCloudSdkGoERROR)
	}

	return resourceAliCloudFCCustomDomainRead(d, meta)
}

func resourceAliCloudFCCustomDomainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	domainName := d.Id()
	objectRaw, err := fcService.DescribeFCCustomDomain(domainName)
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_fc_custom_domain DescribeFCCustomDomain Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Use the service layer helper to set schema fields
	err = fcService.SetSchemaFromCustomDomain(d, objectRaw)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func resourceAliCloudFCCustomDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	domainName := d.Id()

	log.Printf("[DEBUG] Updating FC Custom Domain: %s", domainName)

	// Check if any field has changed
	if d.HasChange("protocol") || d.HasChange("auth_config") || d.HasChange("cert_config") ||
		d.HasChange("route_config") || d.HasChange("tls_config") || d.HasChange("waf_config") {
		// Build update custom domain from schema
		domain := fcService.BuildUpdateCustomDomainInputFromSchema(d)

		// Update custom domain using service layer
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := fcService.UpdateFCCustomDomain(domainName, domain)
			if err != nil {
				if NeedRetry(err) {
					log.Printf("[WARN] FC Custom Domain update failed with retryable error: %s. Retrying...", err)
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				log.Printf("[ERROR] FC Custom Domain update failed: %s", err)
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateCustomDomain", AlibabaCloudSdkGoERROR)
		}

		log.Printf("[DEBUG] FC Custom Domain updated successfully: %s", domainName)
	}

	return resourceAliCloudFCCustomDomainRead(d, meta)
}

func resourceAliCloudFCCustomDomainDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	domainName := d.Id()

	log.Printf("[DEBUG] Deleting FC Custom Domain: %s", domainName)

	// Delete custom domain using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := fcService.DeleteFCCustomDomain(domainName)
		if err != nil {
			if IsNotFoundError(err) {
				log.Printf("[DEBUG] FC Custom Domain not found during deletion: %s", domainName)
				return nil
			}
			if NeedRetry(err) {
				log.Printf("[WARN] FC Custom Domain deletion failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Custom Domain deletion failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteCustomDomain", AlibabaCloudSdkGoERROR)
	}

	log.Printf("[DEBUG] FC Custom Domain deleted successfully: %s", domainName)

	return nil
}
