package alicloud

import (
	"fmt"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

// Custom Domain methods for FCService

// EncodeCustomDomainId encodes domain name into an ID string
func EncodeCustomDomainId(domainName string) string {
	return domainName
}

// DecodeCustomDomainId decodes custom domain ID string to domain name
func DecodeCustomDomainId(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("invalid custom domain ID format, cannot be empty")
	}
	return id, nil
}

// DescribeFCCustomDomain retrieves custom domain information by domain name
func (s *FCService) DescribeFCCustomDomain(domainName string) (*aliyunFCAPI.CustomDomain, error) {
	if domainName == "" {
		return nil, fmt.Errorf("domain name cannot be empty")
	}
	return s.GetAPI().GetCustomDomain(domainName)
}

// ListFCCustomDomains lists all custom domains with optional filters
func (s *FCService) ListFCCustomDomains(prefix *string, limit *int32, nextToken *string) ([]*aliyunFCAPI.CustomDomain, error) {
	var prefixStr, nextTokenStr string
	var limitInt int32

	if prefix != nil {
		prefixStr = *prefix
	}
	if nextToken != nil {
		nextTokenStr = *nextToken
	}
	if limit != nil {
		limitInt = *limit
	}

	domains, _, err := s.GetAPI().ListCustomDomains(prefixStr, nextTokenStr, limitInt)
	return domains, err
}

// CreateFCCustomDomain creates a new FC custom domain
func (s *FCService) CreateFCCustomDomain(domain *aliyunFCAPI.CustomDomain) (*aliyunFCAPI.CustomDomain, error) {
	if domain == nil {
		return nil, fmt.Errorf("domain cannot be nil")
	}
	return s.GetAPI().CreateCustomDomain(domain)
}

// UpdateFCCustomDomain updates an existing FC custom domain
func (s *FCService) UpdateFCCustomDomain(domainName string, domain *aliyunFCAPI.CustomDomain) (*aliyunFCAPI.CustomDomain, error) {
	if domainName == "" {
		return nil, fmt.Errorf("domain name cannot be empty")
	}
	if domain == nil {
		return nil, fmt.Errorf("domain cannot be nil")
	}
	return s.GetAPI().UpdateCustomDomain(domainName, domain)
}

// DeleteFCCustomDomain deletes an FC custom domain
func (s *FCService) DeleteFCCustomDomain(domainName string) error {
	if domainName == "" {
		return fmt.Errorf("domain name cannot be empty")
	}
	return s.GetAPI().DeleteCustomDomain(domainName)
}

// BuildCreateCustomDomainInputFromSchema builds CustomDomain from Terraform schema data
func (s *FCService) BuildCreateCustomDomainInputFromSchema(d *schema.ResourceData) *aliyunFCAPI.CustomDomain {
	domain := &aliyunFCAPI.CustomDomain{}

	if v, ok := d.GetOk("custom_domain_name"); ok {
		domain.DomainName = tea.String(v.(string))
	}

	if v, ok := d.GetOk("protocol"); ok {
		domain.Protocol = tea.String(v.(string))
	}

	// Add auth config
	if v, ok := d.GetOk("auth_config"); ok {
		if authConfigs := v.([]interface{}); len(authConfigs) > 0 {
			authConfig := authConfigs[0].(map[string]interface{})
			domain.AuthConfig = &aliyunFCAPI.AuthConfig{}

			if authType, ok := authConfig["auth_type"].(string); ok && authType != "" {
				domain.AuthConfig.AuthType = tea.String(authType)
			}

			if authInfo, ok := authConfig["auth_info"].(string); ok && authInfo != "" {
				domain.AuthConfig.AuthInfo = tea.String(authInfo)
			}
		}
	}

	// Add cert config
	if v, ok := d.GetOk("cert_config"); ok {
		if certConfigs := v.([]interface{}); len(certConfigs) > 0 {
			certConfig := certConfigs[0].(map[string]interface{})
			domain.CertConfig = &aliyunFCAPI.CertConfig{}

			if certName, ok := certConfig["cert_name"].(string); ok && certName != "" {
				domain.CertConfig.CertName = tea.String(certName)
			}

			if certificate, ok := certConfig["certificate"].(string); ok && certificate != "" {
				domain.CertConfig.Certificate = tea.String(certificate)
			}

			if privateKey, ok := certConfig["private_key"].(string); ok && privateKey != "" {
				domain.CertConfig.PrivateKey = tea.String(privateKey)
			}
		}
	}

	// Add route config
	if v, ok := d.GetOk("route_config"); ok {
		if routeConfigs := v.([]interface{}); len(routeConfigs) > 0 {
			routeConfig := routeConfigs[0].(map[string]interface{})
			domain.RouteConfig = &aliyunFCAPI.RouteConfig{}

			if routes, ok := routeConfig["routes"].([]interface{}); ok && len(routes) > 0 {
				domain.RouteConfig.Routes = make([]*aliyunFCAPI.Route, len(routes))
				for i, route := range routes {
					routeMap := route.(map[string]interface{})
					domainRoute := &aliyunFCAPI.Route{}

					if path, ok := routeMap["path"].(string); ok && path != "" {
						domainRoute.Path = tea.String(path)
					}

					if functionName, ok := routeMap["function_name"].(string); ok && functionName != "" {
						domainRoute.FunctionName = tea.String(functionName)
					}

					if qualifier, ok := routeMap["qualifier"].(string); ok && qualifier != "" {
						domainRoute.Qualifier = tea.String(qualifier)
					}

					if methods, ok := routeMap["methods"].([]interface{}); ok && len(methods) > 0 {
						domainRoute.Methods = make([]*string, len(methods))
						for j, method := range methods {
							domainRoute.Methods[j] = tea.String(method.(string))
						}
					}

					// Add rewrite config
					if rewriteConfigData, ok := routeMap["rewrite_config"].([]interface{}); ok && len(rewriteConfigData) > 0 {
						rewriteConfigMap := rewriteConfigData[0].(map[string]interface{})
						domainRoute.RewriteConfig = &aliyunFCAPI.RewriteConfig{}

						// Add equal rules
						if equalRules, ok := rewriteConfigMap["equal_rules"].([]interface{}); ok && len(equalRules) > 0 {
							domainRoute.RewriteConfig.EqualRules = make([]*aliyunFCAPI.EqualRule, len(equalRules))
							for j, rule := range equalRules {
								ruleMap := rule.(map[string]interface{})
								domainRule := &aliyunFCAPI.EqualRule{}

								if match, ok := ruleMap["match"].(string); ok && match != "" {
									domainRule.Match = tea.String(match)
								}

								if replacement, ok := ruleMap["replacement"].(string); ok && replacement != "" {
									domainRule.Replacement = tea.String(replacement)
								}

								domainRoute.RewriteConfig.EqualRules[j] = domainRule
							}
						}

						// Add regex rules
						if regexRules, ok := rewriteConfigMap["regex_rules"].([]interface{}); ok && len(regexRules) > 0 {
							domainRoute.RewriteConfig.RegexRules = make([]*aliyunFCAPI.RegexRule, len(regexRules))
							for j, rule := range regexRules {
								ruleMap := rule.(map[string]interface{})
								domainRule := &aliyunFCAPI.RegexRule{}

								if match, ok := ruleMap["match"].(string); ok && match != "" {
									domainRule.Match = tea.String(match)
								}

								if replacement, ok := ruleMap["replacement"].(string); ok && replacement != "" {
									domainRule.Replacement = tea.String(replacement)
								}

								domainRoute.RewriteConfig.RegexRules[j] = domainRule
							}
						}

						// Add wildcard rules
						if wildcardRules, ok := rewriteConfigMap["wildcard_rules"].([]interface{}); ok && len(wildcardRules) > 0 {
							domainRoute.RewriteConfig.WildcardRules = make([]*aliyunFCAPI.WildcardRule, len(wildcardRules))
							for j, rule := range wildcardRules {
								ruleMap := rule.(map[string]interface{})
								domainRule := &aliyunFCAPI.WildcardRule{}

								if match, ok := ruleMap["match"].(string); ok && match != "" {
									domainRule.Match = tea.String(match)
								}

								if replacement, ok := ruleMap["replacement"].(string); ok && replacement != "" {
									domainRule.Replacement = tea.String(replacement)
								}

								domainRoute.RewriteConfig.WildcardRules[j] = domainRule
							}
						}
					}

					domain.RouteConfig.Routes[i] = domainRoute
				}
			}
		}
	}

	// Add TLS config
	if v, ok := d.GetOk("tls_config"); ok {
		if tlsConfigs := v.([]interface{}); len(tlsConfigs) > 0 {
			tlsConfig := tlsConfigs[0].(map[string]interface{})
			domain.TlsConfig = &aliyunFCAPI.TLSConfig{}

			if minVersion, ok := tlsConfig["min_version"].(string); ok && minVersion != "" {
				domain.TlsConfig.MinVersion = tea.String(minVersion)
			}

			if maxVersion, ok := tlsConfig["max_version"].(string); ok && maxVersion != "" {
				domain.TlsConfig.MaxVersion = tea.String(maxVersion)
			}

			if cipherSuites, ok := tlsConfig["cipher_suites"].([]interface{}); ok && len(cipherSuites) > 0 {
				domain.TlsConfig.CipherSuites = make([]*string, len(cipherSuites))
				for i, cipher := range cipherSuites {
					domain.TlsConfig.CipherSuites[i] = tea.String(cipher.(string))
				}
			}
		}
	}

	// Add WAF config
	if v, ok := d.GetOk("waf_config"); ok {
		if wafConfigs := v.([]interface{}); len(wafConfigs) > 0 {
			wafConfig := wafConfigs[0].(map[string]interface{})
			domain.WafConfig = &aliyunFCAPI.WAFConfig{}

			if enableWaf, ok := wafConfig["enable_waf"].(bool); ok {
				domain.WafConfig.EnableWAF = tea.Bool(enableWaf)
			}
		}
	}

	return domain
}

// BuildUpdateCustomDomainInputFromSchema builds CustomDomain for update from Terraform schema data
func (s *FCService) BuildUpdateCustomDomainInputFromSchema(d *schema.ResourceData) *aliyunFCAPI.CustomDomain {
	domain := &aliyunFCAPI.CustomDomain{}

	if d.HasChange("protocol") {
		if v, ok := d.GetOk("protocol"); ok {
			domain.Protocol = tea.String(v.(string))
		}
	}

	if d.HasChange("auth_config") {
		if v, ok := d.GetOk("auth_config"); ok {
			if authConfigs := v.([]interface{}); len(authConfigs) > 0 {
				authConfig := authConfigs[0].(map[string]interface{})
				domain.AuthConfig = &aliyunFCAPI.AuthConfig{}

				if authType, ok := authConfig["auth_type"].(string); ok && authType != "" {
					domain.AuthConfig.AuthType = tea.String(authType)
				}

				if authInfo, ok := authConfig["auth_info"].(string); ok && authInfo != "" {
					domain.AuthConfig.AuthInfo = tea.String(authInfo)
				}
			}
		}
	}

	if d.HasChange("cert_config") {
		if v, ok := d.GetOk("cert_config"); ok {
			if certConfigs := v.([]interface{}); len(certConfigs) > 0 {
				certConfig := certConfigs[0].(map[string]interface{})
				domain.CertConfig = &aliyunFCAPI.CertConfig{}

				if certName, ok := certConfig["cert_name"].(string); ok && certName != "" {
					domain.CertConfig.CertName = tea.String(certName)
				}

				if certificate, ok := certConfig["certificate"].(string); ok && certificate != "" {
					domain.CertConfig.Certificate = tea.String(certificate)
				}

				if privateKey, ok := certConfig["private_key"].(string); ok && privateKey != "" {
					domain.CertConfig.PrivateKey = tea.String(privateKey)
				}
			}
		}
	}

	if d.HasChange("route_config") {
		if v, ok := d.GetOk("route_config"); ok {
			if routeConfigs := v.([]interface{}); len(routeConfigs) > 0 {
				routeConfig := routeConfigs[0].(map[string]interface{})
				domain.RouteConfig = &aliyunFCAPI.RouteConfig{}

				if routes, ok := routeConfig["routes"].([]interface{}); ok && len(routes) > 0 {
					domain.RouteConfig.Routes = make([]*aliyunFCAPI.Route, len(routes))
					for i, route := range routes {
						routeMap := route.(map[string]interface{})
						domainRoute := &aliyunFCAPI.Route{}

						if path, ok := routeMap["path"].(string); ok && path != "" {
							domainRoute.Path = tea.String(path)
						}

						if functionName, ok := routeMap["function_name"].(string); ok && functionName != "" {
							domainRoute.FunctionName = tea.String(functionName)
						}

						if qualifier, ok := routeMap["qualifier"].(string); ok && qualifier != "" {
							domainRoute.Qualifier = tea.String(qualifier)
						}

						if methods, ok := routeMap["methods"].([]interface{}); ok && len(methods) > 0 {
							domainRoute.Methods = make([]*string, len(methods))
							for j, method := range methods {
								domainRoute.Methods[j] = tea.String(method.(string))
							}
						}

						// Add rewrite config
						if rewriteConfigData, ok := routeMap["rewrite_config"].([]interface{}); ok && len(rewriteConfigData) > 0 {
							rewriteConfigMap := rewriteConfigData[0].(map[string]interface{})
							domainRoute.RewriteConfig = &aliyunFCAPI.RewriteConfig{}

							// Add equal rules
							if equalRules, ok := rewriteConfigMap["equal_rules"].([]interface{}); ok && len(equalRules) > 0 {
								domainRoute.RewriteConfig.EqualRules = make([]*aliyunFCAPI.EqualRule, len(equalRules))
								for j, rule := range equalRules {
									ruleMap := rule.(map[string]interface{})
									domainRule := &aliyunFCAPI.EqualRule{}

									if match, ok := ruleMap["match"].(string); ok && match != "" {
										domainRule.Match = tea.String(match)
									}

									if replacement, ok := ruleMap["replacement"].(string); ok && replacement != "" {
										domainRule.Replacement = tea.String(replacement)
									}

									domainRoute.RewriteConfig.EqualRules[j] = domainRule
								}
							}

							// Add regex rules
							if regexRules, ok := rewriteConfigMap["regex_rules"].([]interface{}); ok && len(regexRules) > 0 {
								domainRoute.RewriteConfig.RegexRules = make([]*aliyunFCAPI.RegexRule, len(regexRules))
								for j, rule := range regexRules {
									ruleMap := rule.(map[string]interface{})
									domainRule := &aliyunFCAPI.RegexRule{}

									if match, ok := ruleMap["match"].(string); ok && match != "" {
										domainRule.Match = tea.String(match)
									}

									if replacement, ok := ruleMap["replacement"].(string); ok && replacement != "" {
										domainRule.Replacement = tea.String(replacement)
									}

									domainRoute.RewriteConfig.RegexRules[j] = domainRule
								}
							}

							// Add wildcard rules
							if wildcardRules, ok := rewriteConfigMap["wildcard_rules"].([]interface{}); ok && len(wildcardRules) > 0 {
								domainRoute.RewriteConfig.WildcardRules = make([]*aliyunFCAPI.WildcardRule, len(wildcardRules))
								for j, rule := range wildcardRules {
									ruleMap := rule.(map[string]interface{})
									domainRule := &aliyunFCAPI.WildcardRule{}

									if match, ok := ruleMap["match"].(string); ok && match != "" {
										domainRule.Match = tea.String(match)
									}

									if replacement, ok := ruleMap["replacement"].(string); ok && replacement != "" {
										domainRule.Replacement = tea.String(replacement)
									}

									domainRoute.RewriteConfig.WildcardRules[j] = domainRule
								}
							}
						}

						domain.RouteConfig.Routes[i] = domainRoute
					}
				}
			}
		}
	}

	if d.HasChange("tls_config") {
		if v, ok := d.GetOk("tls_config"); ok {
			if tlsConfigs := v.([]interface{}); len(tlsConfigs) > 0 {
				tlsConfig := tlsConfigs[0].(map[string]interface{})
				domain.TlsConfig = &aliyunFCAPI.TLSConfig{}

				if minVersion, ok := tlsConfig["min_version"].(string); ok && minVersion != "" {
					domain.TlsConfig.MinVersion = tea.String(minVersion)
				}

				if maxVersion, ok := tlsConfig["max_version"].(string); ok && maxVersion != "" {
					domain.TlsConfig.MaxVersion = tea.String(maxVersion)
				}

				if cipherSuites, ok := tlsConfig["cipher_suites"].([]interface{}); ok && len(cipherSuites) > 0 {
					domain.TlsConfig.CipherSuites = make([]*string, len(cipherSuites))
					for i, cipher := range cipherSuites {
						domain.TlsConfig.CipherSuites[i] = tea.String(cipher.(string))
					}
				}
			}
		}
	}

	if d.HasChange("waf_config") {
		if v, ok := d.GetOk("waf_config"); ok {
			if wafConfigs := v.([]interface{}); len(wafConfigs) > 0 {
				wafConfig := wafConfigs[0].(map[string]interface{})
				domain.WafConfig = &aliyunFCAPI.WAFConfig{}

				if enableWaf, ok := wafConfig["enable_waf"].(bool); ok {
					domain.WafConfig.EnableWAF = tea.Bool(enableWaf)
				}
			}
		}
	}

	return domain
}

// SetSchemaFromCustomDomain sets terraform schema data from CustomDomain
func (s *FCService) SetSchemaFromCustomDomain(d *schema.ResourceData, domain *aliyunFCAPI.CustomDomain) error {
	if domain == nil {
		return fmt.Errorf("domain cannot be nil")
	}

	if domain.DomainName != nil {
		d.Set("custom_domain_name", *domain.DomainName)
	}

	if domain.Protocol != nil {
		d.Set("protocol", *domain.Protocol)
	}

	if domain.AccountId != nil {
		d.Set("account_id", *domain.AccountId)
	}

	if domain.CreatedTime != nil {
		d.Set("create_time", *domain.CreatedTime)
	}

	if domain.LastModifiedTime != nil {
		d.Set("last_modified_time", *domain.LastModifiedTime)
	}

	if domain.SubdomainCount != nil {
		d.Set("subdomain_count", *domain.SubdomainCount)
	}

	// Set auth config
	if domain.AuthConfig != nil {
		authConfigMaps := make([]map[string]interface{}, 0)
		authConfigMap := make(map[string]interface{})

		if domain.AuthConfig.AuthType != nil {
			authConfigMap["auth_type"] = *domain.AuthConfig.AuthType
		}

		if domain.AuthConfig.AuthInfo != nil {
			authConfigMap["auth_info"] = *domain.AuthConfig.AuthInfo
		}

		if len(authConfigMap) > 0 {
			authConfigMaps = append(authConfigMaps, authConfigMap)
			d.Set("auth_config", authConfigMaps)
		}
	}

	// Set cert config
	if domain.CertConfig != nil {
		certConfigMaps := make([]map[string]interface{}, 0)
		certConfigMap := make(map[string]interface{})

		if domain.CertConfig.CertName != nil {
			certConfigMap["cert_name"] = *domain.CertConfig.CertName
		}

		if domain.CertConfig.Certificate != nil {
			certConfigMap["certificate"] = *domain.CertConfig.Certificate
		}

		// The FC service will not return private key credential for security reason.
		// Read it from the terraform file.
		oldConfig := d.Get("cert_config").([]interface{})
		if len(oldConfig) > 0 {
			certConfigMap["private_key"] = oldConfig[0].(map[string]interface{})["private_key"]
		}

		if len(certConfigMap) > 0 {
			certConfigMaps = append(certConfigMaps, certConfigMap)
			d.Set("cert_config", certConfigMaps)
		}
	}

	// Set route config
	if domain.RouteConfig != nil {
		routeConfigMaps := make([]map[string]interface{}, 0)
		routeConfigMap := make(map[string]interface{})

		if domain.RouteConfig.Routes != nil {
			routesMaps := make([]map[string]interface{}, 0)

			for _, route := range domain.RouteConfig.Routes {
				if route != nil {
					routesMap := make(map[string]interface{})

					if route.Path != nil {
						routesMap["path"] = *route.Path
					}

					if route.FunctionName != nil {
						routesMap["function_name"] = *route.FunctionName
					}

					if route.Qualifier != nil {
						routesMap["qualifier"] = *route.Qualifier
					}

					// Set methods
					if route.Methods != nil {
						methods := make([]interface{}, len(route.Methods))
						for i, method := range route.Methods {
							if method != nil {
								methods[i] = *method
							}
						}
						routesMap["methods"] = methods
					}

					// Set rewrite config
					if route.RewriteConfig != nil {
						rewriteConfigMaps := make([]map[string]interface{}, 0)
						rewriteConfigMap := make(map[string]interface{})

						// Set equal rules
						if route.RewriteConfig.EqualRules != nil {
							equalRulesMaps := make([]map[string]interface{}, 0)
							for _, rule := range route.RewriteConfig.EqualRules {
								if rule != nil {
									equalRuleMap := make(map[string]interface{})

									if rule.Match != nil {
										equalRuleMap["match"] = *rule.Match
									}

									if rule.Replacement != nil {
										equalRuleMap["replacement"] = *rule.Replacement
									}

									equalRulesMaps = append(equalRulesMaps, equalRuleMap)
								}
							}
							rewriteConfigMap["equal_rules"] = equalRulesMaps
						}

						// Set regex rules
						if route.RewriteConfig.RegexRules != nil {
							regexRulesMaps := make([]map[string]interface{}, 0)
							for _, rule := range route.RewriteConfig.RegexRules {
								if rule != nil {
									regexRuleMap := make(map[string]interface{})

									if rule.Match != nil {
										regexRuleMap["match"] = *rule.Match
									}

									if rule.Replacement != nil {
										regexRuleMap["replacement"] = *rule.Replacement
									}

									regexRulesMaps = append(regexRulesMaps, regexRuleMap)
								}
							}
							rewriteConfigMap["regex_rules"] = regexRulesMaps
						}

						// Set wildcard rules
						if route.RewriteConfig.WildcardRules != nil {
							wildcardRulesMaps := make([]map[string]interface{}, 0)
							for _, rule := range route.RewriteConfig.WildcardRules {
								if rule != nil {
									wildcardRuleMap := make(map[string]interface{})

									if rule.Match != nil {
										wildcardRuleMap["match"] = *rule.Match
									}

									if rule.Replacement != nil {
										wildcardRuleMap["replacement"] = *rule.Replacement
									}

									wildcardRulesMaps = append(wildcardRulesMaps, wildcardRuleMap)
								}
							}
							rewriteConfigMap["wildcard_rules"] = wildcardRulesMaps
						}

						if len(rewriteConfigMap) > 0 {
							rewriteConfigMaps = append(rewriteConfigMaps, rewriteConfigMap)
							routesMap["rewrite_config"] = rewriteConfigMaps
						}
					}

					routesMaps = append(routesMaps, routesMap)
				}
			}

			routeConfigMap["routes"] = routesMaps
		}

		if len(routeConfigMap) > 0 {
			routeConfigMaps = append(routeConfigMaps, routeConfigMap)
			d.Set("route_config", routeConfigMaps)
		}
	}

	// Set TLS config
	if domain.TlsConfig != nil {
		tlsConfigMaps := make([]map[string]interface{}, 0)
		tlsConfigMap := make(map[string]interface{})

		if domain.TlsConfig.MinVersion != nil {
			tlsConfigMap["min_version"] = *domain.TlsConfig.MinVersion
		}

		if domain.TlsConfig.MaxVersion != nil {
			tlsConfigMap["max_version"] = *domain.TlsConfig.MaxVersion
		}

		// Set cipher suites
		if domain.TlsConfig.CipherSuites != nil {
			cipherSuites := make([]interface{}, len(domain.TlsConfig.CipherSuites))
			for i, cipher := range domain.TlsConfig.CipherSuites {
				if cipher != nil {
					cipherSuites[i] = *cipher
				}
			}
			tlsConfigMap["cipher_suites"] = cipherSuites
		}

		if len(tlsConfigMap) > 0 {
			tlsConfigMaps = append(tlsConfigMaps, tlsConfigMap)
			d.Set("tls_config", tlsConfigMaps)
		}
	}

	// Set WAF config
	if domain.WafConfig != nil {
		wafConfigMaps := make([]map[string]interface{}, 0)
		wafConfigMap := make(map[string]interface{})

		if domain.WafConfig.EnableWAF != nil {
			wafConfigMap["enable_waf"] = *domain.WafConfig.EnableWAF
		}

		if len(wafConfigMap) > 0 {
			wafConfigMaps = append(wafConfigMaps, wafConfigMap)
			d.Set("waf_config", wafConfigMaps)
		}
	}

	return nil
}

// CustomDomainStateRefreshFunc returns a StateRefreshFunc to wait for custom domain status changes
func (s *FCService) CustomDomainStateRefreshFunc(domainName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFCCustomDomain(domainName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		currentState := "Active" // FC v3 custom domains are typically Active when created
		if object.DomainName != nil && *object.DomainName != "" {
			// Custom domains exist, so they are active
			currentState = "Active"
		}

		for _, failState := range failStates {
			if currentState == failState {
				return object, currentState, WrapError(Error(FailedToReachTargetStatus, currentState))
			}
		}
		return object, currentState, nil
	}
}

// WaitForCustomDomainCreating waits for custom domain creation to complete
func (s *FCService) WaitForCustomDomainCreating(domainName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.CustomDomainStateRefreshFunc(domainName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, domainName)
}

// WaitForCustomDomainDeleting waits for custom domain deletion to complete
func (s *FCService) WaitForCustomDomainDeleting(domainName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Deleting", "Active"},
		[]string{""},
		timeout,
		5*time.Second,
		s.CustomDomainStateRefreshFunc(domainName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, domainName)
}

// WaitForCustomDomainUpdating waits for custom domain update to complete
func (s *FCService) WaitForCustomDomainUpdating(domainName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Updating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.CustomDomainStateRefreshFunc(domainName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, domainName)
}
