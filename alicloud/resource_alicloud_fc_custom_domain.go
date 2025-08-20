package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
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

	action := fmt.Sprintf("/2023-03-30/custom-domains")
	var request map[string]interface{}
	var response map[string]interface{}
	query := make(map[string]*string)
	body := make(map[string]interface{})
	var err error
	request = make(map[string]interface{})
	if v, ok := d.GetOk("custom_domain_name"); ok {
		request["domainName"] = v
	}

	objectDataLocalMap := make(map[string]interface{})

	if v := d.Get("auth_config"); !IsNil(v) {
		authInfo1, _ := jsonpath.Get("$[0].auth_info", d.Get("auth_config"))
		if authInfo1 != nil && authInfo1 != "" {
			objectDataLocalMap["authInfo"] = authInfo1
		}
		authType1, _ := jsonpath.Get("$[0].auth_type", d.Get("auth_config"))
		if authType1 != nil && authType1 != "" {
			objectDataLocalMap["authType"] = authType1
		}

		request["authConfig"] = objectDataLocalMap
	}

	objectDataLocalMap1 := make(map[string]interface{})

	if v := d.Get("cert_config"); !IsNil(v) {
		certName1, _ := jsonpath.Get("$[0].cert_name", d.Get("cert_config"))
		if certName1 != nil && certName1 != "" {
			objectDataLocalMap1["certName"] = certName1
		}
		certificate1, _ := jsonpath.Get("$[0].certificate", d.Get("cert_config"))
		if certificate1 != nil && certificate1 != "" {
			objectDataLocalMap1["certificate"] = certificate1
		}
		privateKey1, _ := jsonpath.Get("$[0].private_key", d.Get("cert_config"))
		if privateKey1 != nil && privateKey1 != "" {
			objectDataLocalMap1["privateKey"] = privateKey1
		}

		request["certConfig"] = objectDataLocalMap1
	}

	if v, ok := d.GetOk("protocol"); ok {
		request["protocol"] = v
	}
	objectDataLocalMap2 := make(map[string]interface{})

	if v := d.Get("tls_config"); !IsNil(v) {
		cipherSuites1, _ := jsonpath.Get("$[0].cipher_suites", v)
		if cipherSuites1 != nil && cipherSuites1 != "" {
			objectDataLocalMap2["cipherSuites"] = cipherSuites1
		}
		maxVersion1, _ := jsonpath.Get("$[0].max_version", d.Get("tls_config"))
		if maxVersion1 != nil && maxVersion1 != "" {
			objectDataLocalMap2["maxVersion"] = maxVersion1
		}
		minVersion1, _ := jsonpath.Get("$[0].min_version", d.Get("tls_config"))
		if minVersion1 != nil && minVersion1 != "" {
			objectDataLocalMap2["minVersion"] = minVersion1
		}

		request["tlsConfig"] = objectDataLocalMap2
	}

	objectDataLocalMap3 := make(map[string]interface{})

	if v := d.Get("route_config"); !IsNil(v) {
		if v, ok := d.GetOk("route_config"); ok {
			localData, err := jsonpath.Get("$[0].routes", v)
			if err != nil {
				localData = make([]interface{}, 0)
			}
			localMaps := make([]interface{}, 0)
			for _, dataLoop := range localData.([]interface{}) {
				dataLoopTmp := make(map[string]interface{})
				if dataLoop != nil {
					dataLoopTmp = dataLoop.(map[string]interface{})
				}
				dataLoopMap := make(map[string]interface{})
				dataLoopMap["methods"] = dataLoopTmp["methods"]
				dataLoopMap["functionName"] = dataLoopTmp["function_name"]
				dataLoopMap["path"] = dataLoopTmp["path"]
				dataLoopMap["qualifier"] = dataLoopTmp["qualifier"]
				localData1 := make(map[string]interface{})
				if v, ok := dataLoopTmp["rewrite_config"]; ok {
					localData2, err := jsonpath.Get("$[0].equal_rules", v)
					if err != nil {
						localData2 = make([]interface{}, 0)
					}
					localMaps2 := make([]interface{}, 0)
					for _, dataLoop2 := range localData2.([]interface{}) {
						dataLoop2Tmp := make(map[string]interface{})
						if dataLoop2 != nil {
							dataLoop2Tmp = dataLoop2.(map[string]interface{})
						}
						dataLoop2Map := make(map[string]interface{})
						dataLoop2Map["match"] = dataLoop2Tmp["match"]
						dataLoop2Map["replacement"] = dataLoop2Tmp["replacement"]
						localMaps2 = append(localMaps2, dataLoop2Map)
					}
					localData1["equalRules"] = localMaps2
				}

				if v, ok := dataLoopTmp["rewrite_config"]; ok {
					localData3, err := jsonpath.Get("$[0].regex_rules", v)
					if err != nil {
						localData3 = make([]interface{}, 0)
					}
					localMaps3 := make([]interface{}, 0)
					for _, dataLoop3 := range localData3.([]interface{}) {
						dataLoop3Tmp := make(map[string]interface{})
						if dataLoop3 != nil {
							dataLoop3Tmp = dataLoop3.(map[string]interface{})
						}
						dataLoop3Map := make(map[string]interface{})
						dataLoop3Map["match"] = dataLoop3Tmp["match"]
						dataLoop3Map["replacement"] = dataLoop3Tmp["replacement"]
						localMaps3 = append(localMaps3, dataLoop3Map)
					}
					localData1["regexRules"] = localMaps3
				}

				if v, ok := dataLoopTmp["rewrite_config"]; ok {
					localData4, err := jsonpath.Get("$[0].wildcard_rules", v)
					if err != nil {
						localData4 = make([]interface{}, 0)
					}
					localMaps4 := make([]interface{}, 0)
					for _, dataLoop4 := range localData4.([]interface{}) {
						dataLoop4Tmp := make(map[string]interface{})
						if dataLoop4 != nil {
							dataLoop4Tmp = dataLoop4.(map[string]interface{})
						}
						dataLoop4Map := make(map[string]interface{})
						dataLoop4Map["match"] = dataLoop4Tmp["match"]
						dataLoop4Map["replacement"] = dataLoop4Tmp["replacement"]
						localMaps4 = append(localMaps4, dataLoop4Map)
					}
					localData1["wildcardRules"] = localMaps4
				}

				dataLoopMap["rewriteConfig"] = localData1
				localMaps = append(localMaps, dataLoopMap)
			}
			objectDataLocalMap3["routes"] = localMaps
		}

		request["routeConfig"] = objectDataLocalMap3
	}

	objectDataLocalMap4 := make(map[string]interface{})

	if v := d.Get("waf_config"); !IsNil(v) {
		enableWaf, _ := jsonpath.Get("$[0].enable_waf", d.Get("waf_config"))
		if enableWaf != nil && enableWaf != "" {
			objectDataLocalMap4["enableWAF"] = enableWaf
		}

		request["wafConfig"] = objectDataLocalMap4
	}

	body = request
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err = client.RoaPost("FC", "2023-03-30", action, query, nil, body, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_fc_custom_domain", action, AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprint(response["domainName"]))

	return resourceAliCloudFCCustomDomainRead(d, meta)
}

func resourceAliCloudFCCustomDomainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	objectRaw, err := fcService.DescribeFCCustomDomain(d.Id())
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_fc_custom_domain DescribeFCCustomDomain Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if objectRaw.AccountId != nil && *objectRaw.AccountId != "" {
		d.Set("account_id", *objectRaw.AccountId)
	}
	// Note: API version is not available in FC v3 custom domain
	if objectRaw.CreatedTime != nil {
		d.Set("create_time", *objectRaw.CreatedTime)
	}
	if objectRaw.LastModifiedTime != nil {
		d.Set("last_modified_time", *objectRaw.LastModifiedTime)
	}
	if objectRaw.Protocol != nil && *objectRaw.Protocol != "" {
		d.Set("protocol", *objectRaw.Protocol)
	}
	if objectRaw.SubdomainCount != nil {
		d.Set("subdomain_count", *objectRaw.SubdomainCount)
	}
	if objectRaw.DomainName != nil {
		d.Set("custom_domain_name", *objectRaw.DomainName)
	}

	// AuthConfig
	authConfigMaps := make([]map[string]interface{}, 0)
	if objectRaw.AuthConfig != nil {
		authConfigMap := make(map[string]interface{})
		if objectRaw.AuthConfig.AuthInfo != nil {
			authConfigMap["auth_info"] = *objectRaw.AuthConfig.AuthInfo
		}
		if objectRaw.AuthConfig.AuthType != nil {
			authConfigMap["auth_type"] = *objectRaw.AuthConfig.AuthType
		}
		if len(authConfigMap) > 0 {
			authConfigMaps = append(authConfigMaps, authConfigMap)
		}
		if err := d.Set("auth_config", authConfigMaps); err != nil {
			return err
		}
	}

	// CertConfig
	certConfigMaps := make([]map[string]interface{}, 0)
	if objectRaw.CertConfig != nil {
		certConfigMap := make(map[string]interface{})
		if objectRaw.CertConfig.CertName != nil {
			certConfigMap["cert_name"] = *objectRaw.CertConfig.CertName
		}
		if objectRaw.CertConfig.Certificate != nil {
			certConfigMap["certificate"] = *objectRaw.CertConfig.Certificate
		}
		// The FC service will not return private key credential for security reason.
		// Read it from the terraform file.
		oldConfig := d.Get("cert_config").([]interface{})
		if len(oldConfig) > 0 {
			certConfigMap["private_key"] = oldConfig[0].(map[string]interface{})["private_key"]
		}
		if len(certConfigMap) > 0 {
			certConfigMaps = append(certConfigMaps, certConfigMap)
		}
		if err := d.Set("cert_config", certConfigMaps); err != nil {
			return err
		}
	}

	// RouteConfig
	routeConfigMaps := make([]map[string]interface{}, 0)
	if objectRaw.RouteConfig != nil {
		routeConfigMap := make(map[string]interface{})
		routesMaps := make([]map[string]interface{}, 0)

		if objectRaw.RouteConfig.Routes != nil {
			for _, route := range objectRaw.RouteConfig.Routes {
				if route != nil {
					routesMap := make(map[string]interface{})
					if route.FunctionName != nil {
						routesMap["function_name"] = *route.FunctionName
					}
					if route.Path != nil {
						routesMap["path"] = *route.Path
					}
					if route.Qualifier != nil {
						routesMap["qualifier"] = *route.Qualifier
					}

					// Methods
					methods1Raw := make([]interface{}, 0)
					if route.Methods != nil {
						for _, method := range route.Methods {
							if method != nil {
								methods1Raw = append(methods1Raw, *method)
							}
						}
					}
					routesMap["methods"] = methods1Raw

					// RewriteConfig
					rewriteConfigMaps := make([]map[string]interface{}, 0)
					if route.RewriteConfig != nil {
						rewriteConfigMap := make(map[string]interface{})

						// Equal rules
						equalRulesMaps := make([]map[string]interface{}, 0)
						if route.RewriteConfig.EqualRules != nil {
							for _, equalRule := range route.RewriteConfig.EqualRules {
								if equalRule != nil {
									equalRulesMap := make(map[string]interface{})
									if equalRule.Match != nil {
										equalRulesMap["match"] = *equalRule.Match
									}
									if equalRule.Replacement != nil {
										equalRulesMap["replacement"] = *equalRule.Replacement
									}
									equalRulesMaps = append(equalRulesMaps, equalRulesMap)
								}
							}
						}
						rewriteConfigMap["equal_rules"] = equalRulesMaps

						// Regex rules
						regexRulesMaps := make([]map[string]interface{}, 0)
						if route.RewriteConfig.RegexRules != nil {
							for _, regexRule := range route.RewriteConfig.RegexRules {
								if regexRule != nil {
									regexRulesMap := make(map[string]interface{})
									if regexRule.Match != nil {
										regexRulesMap["match"] = *regexRule.Match
									}
									if regexRule.Replacement != nil {
										regexRulesMap["replacement"] = *regexRule.Replacement
									}
									regexRulesMaps = append(regexRulesMaps, regexRulesMap)
								}
							}
						}
						rewriteConfigMap["regex_rules"] = regexRulesMaps

						// Wildcard rules
						wildcardRulesMaps := make([]map[string]interface{}, 0)
						if route.RewriteConfig.WildcardRules != nil {
							for _, wildcardRule := range route.RewriteConfig.WildcardRules {
								if wildcardRule != nil {
									wildcardRulesMap := make(map[string]interface{})
									if wildcardRule.Match != nil {
										wildcardRulesMap["match"] = *wildcardRule.Match
									}
									if wildcardRule.Replacement != nil {
										wildcardRulesMap["replacement"] = *wildcardRule.Replacement
									}
									wildcardRulesMaps = append(wildcardRulesMaps, wildcardRulesMap)
								}
							}
						}
						rewriteConfigMap["wildcard_rules"] = wildcardRulesMaps

						rewriteConfigMaps = append(rewriteConfigMaps, rewriteConfigMap)
					}
					routesMap["rewrite_config"] = rewriteConfigMaps
					routesMaps = append(routesMaps, routesMap)
				}
			}
		}

		routeConfigMap["routes"] = routesMaps
		routeConfigMaps = append(routeConfigMaps, routeConfigMap)
		if err := d.Set("route_config", routeConfigMaps); err != nil {
			return err
		}
	}

	// TLSConfig
	tlsConfigMaps := make([]map[string]interface{}, 0)
	if objectRaw.TlsConfig != nil {
		tlsConfigMap := make(map[string]interface{})
		if objectRaw.TlsConfig.MaxVersion != nil {
			tlsConfigMap["max_version"] = *objectRaw.TlsConfig.MaxVersion
		}
		if objectRaw.TlsConfig.MinVersion != nil {
			tlsConfigMap["min_version"] = *objectRaw.TlsConfig.MinVersion
		}

		cipherSuites1Raw := make([]interface{}, 0)
		if objectRaw.TlsConfig.CipherSuites != nil {
			for _, cipher := range objectRaw.TlsConfig.CipherSuites {
				if cipher != nil {
					cipherSuites1Raw = append(cipherSuites1Raw, *cipher)
				}
			}
		}
		tlsConfigMap["cipher_suites"] = cipherSuites1Raw

		if len(tlsConfigMap) > 0 {
			tlsConfigMaps = append(tlsConfigMaps, tlsConfigMap)
		}
		if err := d.Set("tls_config", tlsConfigMaps); err != nil {
			return err
		}
	}

	// WAFConfig
	wafConfigMaps := make([]map[string]interface{}, 0)
	if objectRaw.WafConfig != nil {
		wafConfigMap := make(map[string]interface{})
		if objectRaw.WafConfig.EnableWAF != nil {
			wafConfigMap["enable_waf"] = *objectRaw.WafConfig.EnableWAF
		}
		if len(wafConfigMap) > 0 {
			wafConfigMaps = append(wafConfigMaps, wafConfigMap)
		}
		if err := d.Set("waf_config", wafConfigMaps); err != nil {
			return err
		}
	}

	d.Set("custom_domain_name", d.Id())

	return nil
}

func resourceAliCloudFCCustomDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	var body map[string]interface{}
	update := false
	domainName := d.Id()
	action := fmt.Sprintf("/2023-03-30/custom-domains/%s", domainName)
	var err error
	request = make(map[string]interface{})
	query = make(map[string]*string)
	body = make(map[string]interface{})
	request["domainName"] = d.Id()

	if d.HasChange("auth_config") {
		update = true
		objectDataLocalMap := make(map[string]interface{})

		if v := d.Get("auth_config"); !IsNil(v) {
			authInfo1, _ := jsonpath.Get("$[0].auth_info", v)
			if authInfo1 != nil && (d.HasChange("auth_config.0.auth_info") || authInfo1 != "") {
				objectDataLocalMap["authInfo"] = authInfo1
			}
			authType1, _ := jsonpath.Get("$[0].auth_type", v)
			if authType1 != nil && (d.HasChange("auth_config.0.auth_type") || authType1 != "") {
				objectDataLocalMap["authType"] = authType1
			}

			request["authConfig"] = objectDataLocalMap
		}
	}

	if d.HasChange("cert_config") {
		update = true
		objectDataLocalMap1 := make(map[string]interface{})

		if v := d.Get("cert_config"); !IsNil(v) {
			certName1, _ := jsonpath.Get("$[0].cert_name", v)
			if certName1 != nil && (d.HasChange("cert_config.0.cert_name") || certName1 != "") {
				objectDataLocalMap1["certName"] = certName1
			}
			certificate1, _ := jsonpath.Get("$[0].certificate", v)
			if certificate1 != nil && (d.HasChange("cert_config.0.certificate") || certificate1 != "") {
				objectDataLocalMap1["certificate"] = certificate1
			}
			privateKey1, _ := jsonpath.Get("$[0].private_key", v)
			if privateKey1 != nil && (d.HasChange("cert_config.0.private_key") || privateKey1 != "") {
				objectDataLocalMap1["privateKey"] = privateKey1
			}

			request["certConfig"] = objectDataLocalMap1
		}
	}

	if d.HasChange("protocol") {
		update = true
		request["protocol"] = d.Get("protocol")
	}

	if d.HasChange("tls_config") {
		update = true
		objectDataLocalMap2 := make(map[string]interface{})

		if v := d.Get("tls_config"); !IsNil(v) {
			cipherSuites1, _ := jsonpath.Get("$[0].cipher_suites", d.Get("tls_config"))
			if cipherSuites1 != nil && (d.HasChange("tls_config.0.cipher_suites") || cipherSuites1 != "") {
				objectDataLocalMap2["cipherSuites"] = cipherSuites1
			}
			maxVersion1, _ := jsonpath.Get("$[0].max_version", v)
			if maxVersion1 != nil && (d.HasChange("tls_config.0.max_version") || maxVersion1 != "") {
				objectDataLocalMap2["maxVersion"] = maxVersion1
			}
			minVersion1, _ := jsonpath.Get("$[0].min_version", v)
			if minVersion1 != nil && (d.HasChange("tls_config.0.min_version") || minVersion1 != "") {
				objectDataLocalMap2["minVersion"] = minVersion1
			}

			request["tlsConfig"] = objectDataLocalMap2
		}
	}

	if d.HasChange("route_config") {
		update = true
		objectDataLocalMap3 := make(map[string]interface{})

		if v := d.Get("route_config"); !IsNil(v) {
			if v, ok := d.GetOk("route_config"); ok {
				localData, err := jsonpath.Get("$[0].routes", v)
				if err != nil {
					localData = make([]interface{}, 0)
				}
				localMaps := make([]interface{}, 0)
				for _, dataLoop := range localData.([]interface{}) {
					dataLoopTmp := make(map[string]interface{})
					if dataLoop != nil {
						dataLoopTmp = dataLoop.(map[string]interface{})
					}
					dataLoopMap := make(map[string]interface{})
					dataLoopMap["methods"] = dataLoopTmp["methods"]
					dataLoopMap["functionName"] = dataLoopTmp["function_name"]
					dataLoopMap["path"] = dataLoopTmp["path"]
					dataLoopMap["qualifier"] = dataLoopTmp["qualifier"]
					if !IsNil(dataLoopTmp["rewrite_config"]) {
						localData1 := make(map[string]interface{})
						if v, ok := dataLoopTmp["rewrite_config"]; ok {
							localData2, err := jsonpath.Get("$[0].equal_rules", v)
							if err != nil {
								localData2 = make([]interface{}, 0)
							}
							localMaps2 := make([]interface{}, 0)
							for _, dataLoop2 := range localData2.([]interface{}) {
								dataLoop2Tmp := make(map[string]interface{})
								if dataLoop2 != nil {
									dataLoop2Tmp = dataLoop2.(map[string]interface{})
								}
								dataLoop2Map := make(map[string]interface{})
								dataLoop2Map["match"] = dataLoop2Tmp["match"]
								dataLoop2Map["replacement"] = dataLoop2Tmp["replacement"]
								localMaps2 = append(localMaps2, dataLoop2Map)
							}
							localData1["equalRules"] = localMaps2
						}

						if v, ok := dataLoopTmp["rewrite_config"]; ok {
							localData3, err := jsonpath.Get("$[0].regex_rules", v)
							if err != nil {
								localData3 = make([]interface{}, 0)
							}
							localMaps3 := make([]interface{}, 0)
							for _, dataLoop3 := range localData3.([]interface{}) {
								dataLoop3Tmp := make(map[string]interface{})
								if dataLoop3 != nil {
									dataLoop3Tmp = dataLoop3.(map[string]interface{})
								}
								dataLoop3Map := make(map[string]interface{})
								dataLoop3Map["match"] = dataLoop3Tmp["match"]
								dataLoop3Map["replacement"] = dataLoop3Tmp["replacement"]
								localMaps3 = append(localMaps3, dataLoop3Map)
							}
							localData1["regexRules"] = localMaps3
						}

						if v, ok := dataLoopTmp["rewrite_config"]; ok {
							localData4, err := jsonpath.Get("$[0].wildcard_rules", v)
							if err != nil {
								localData4 = make([]interface{}, 0)
							}
							localMaps4 := make([]interface{}, 0)
							for _, dataLoop4 := range localData4.([]interface{}) {
								dataLoop4Tmp := make(map[string]interface{})
								if dataLoop4 != nil {
									dataLoop4Tmp = dataLoop4.(map[string]interface{})
								}
								dataLoop4Map := make(map[string]interface{})
								dataLoop4Map["match"] = dataLoop4Tmp["match"]
								dataLoop4Map["replacement"] = dataLoop4Tmp["replacement"]
								localMaps4 = append(localMaps4, dataLoop4Map)
							}
							localData1["wildcardRules"] = localMaps4
						}

						dataLoopMap["rewriteConfig"] = localData1
					}
					localMaps = append(localMaps, dataLoopMap)
				}
				objectDataLocalMap3["routes"] = localMaps
			}

			request["routeConfig"] = objectDataLocalMap3
		}
	}

	if d.HasChange("waf_config") {
		update = true
		objectDataLocalMap4 := make(map[string]interface{})

		if v := d.Get("waf_config"); !IsNil(v) {
			enableWaf, _ := jsonpath.Get("$[0].enable_waf", v)
			if enableWaf != nil && (d.HasChange("waf_config.0.enable_waf") || enableWaf != "") {
				objectDataLocalMap4["enableWAF"] = enableWaf
			}

			request["wafConfig"] = objectDataLocalMap4
		}
	}

	body = request
	if update {
		wait := incrementalWait(3*time.Second, 5*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err = client.RoaPut("FC", "2023-03-30", action, query, nil, body, true)
			if err != nil {
				if NeedRetry(err) {
					wait()
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		addDebug(action, response, request)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudFCCustomDomainRead(d, meta)
}

func resourceAliCloudFCCustomDomainDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	domainName := d.Id()
	action := fmt.Sprintf("/2023-03-30/custom-domains/%s", domainName)
	var request map[string]interface{}
	var response map[string]interface{}
	query := make(map[string]*string)
	var err error
	request = make(map[string]interface{})
	request["domainName"] = d.Id()

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err = client.RoaDelete("FC", "2023-03-30", action, query, nil, nil, true)

		if err != nil {
			if IsExpectedErrors(err, []string{"429"}) || NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		if IsExpectedErrors(err, []string{"DomainNameNotFound"}) || IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}

	return nil
}
