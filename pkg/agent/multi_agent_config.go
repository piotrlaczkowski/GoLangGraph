// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package agent

import (
	"fmt"
	"strings"
	"time"
)

// MultiAgentConfig represents a configuration for multiple agents
type MultiAgentConfig struct {
	Name        string                  `json:"name" yaml:"name"`
	Version     string                  `json:"version" yaml:"version"`
	Description string                  `json:"description" yaml:"description"`
	Agents      map[string]*AgentConfig `json:"agents" yaml:"agents"`
	Routing     *RoutingConfig          `json:"routing" yaml:"routing"`
	Deployment  *DeploymentConfig       `json:"deployment" yaml:"deployment"`
	Shared      *SharedConfig           `json:"shared" yaml:"shared"`
	Metadata    map[string]interface{}  `json:"metadata" yaml:"metadata"`
}

// RoutingConfig defines how requests are routed to different agents
type RoutingConfig struct {
	Type         string             `json:"type" yaml:"type"` // "path", "header", "query", "subdomain"
	DefaultAgent string             `json:"default_agent" yaml:"default_agent"`
	Rules        []RoutingRule      `json:"rules" yaml:"rules"`
	Middleware   []MiddlewareConfig `json:"middleware" yaml:"middleware"`
}

// RoutingRule defines a routing rule
type RoutingRule struct {
	ID         string                 `json:"id" yaml:"id"`
	Pattern    string                 `json:"pattern" yaml:"pattern"`
	AgentID    string                 `json:"agent_id" yaml:"agent_id"`
	Method     string                 `json:"method" yaml:"method"`
	Priority   int                    `json:"priority" yaml:"priority"`
	Conditions []RoutingCondition     `json:"conditions" yaml:"conditions"`
	Metadata   map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// RoutingCondition defines conditions for routing
type RoutingCondition struct {
	Type     string `json:"type" yaml:"type"` // "header", "query", "body", "ip"
	Key      string `json:"key" yaml:"key"`
	Value    string `json:"value" yaml:"value"`
	Operator string `json:"operator" yaml:"operator"` // "equals", "contains", "regex", "prefix", "suffix"
}

// MiddlewareConfig defines middleware configuration
type MiddlewareConfig struct {
	Type    string                 `json:"type" yaml:"type"`
	Config  map[string]interface{} `json:"config" yaml:"config"`
	Enabled bool                   `json:"enabled" yaml:"enabled"`
}

// DeploymentConfig defines deployment configuration
type DeploymentConfig struct {
	Type        string                 `json:"type" yaml:"type"` // "docker", "kubernetes", "serverless"
	Environment string                 `json:"environment" yaml:"environment"`
	Replicas    int                    `json:"replicas" yaml:"replicas"`
	Resources   *ResourceConfig        `json:"resources" yaml:"resources"`
	Networking  *NetworkingConfig      `json:"networking" yaml:"networking"`
	Scaling     *ScalingConfig         `json:"scaling" yaml:"scaling"`
	HealthCheck *HealthCheckConfig     `json:"health_check" yaml:"health_check"`
	Secrets     map[string]string      `json:"secrets" yaml:"secrets"`
	ConfigMaps  map[string]string      `json:"config_maps" yaml:"config_maps"`
	Volumes     []VolumeConfig         `json:"volumes" yaml:"volumes"`
	Metadata    map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// ResourceConfig defines resource limits and requests
type ResourceConfig struct {
	CPU      string `json:"cpu" yaml:"cpu"`
	Memory   string `json:"memory" yaml:"memory"`
	Storage  string `json:"storage" yaml:"storage"`
	GPU      string `json:"gpu" yaml:"gpu"`
	Requests *struct {
		CPU     string `json:"cpu" yaml:"cpu"`
		Memory  string `json:"memory" yaml:"memory"`
		Storage string `json:"storage" yaml:"storage"`
	} `json:"requests" yaml:"requests"`
	Limits *struct {
		CPU     string `json:"cpu" yaml:"cpu"`
		Memory  string `json:"memory" yaml:"memory"`
		Storage string `json:"storage" yaml:"storage"`
	} `json:"limits" yaml:"limits"`
}

// NetworkingConfig defines networking configuration
type NetworkingConfig struct {
	Type        string            `json:"type" yaml:"type"` // "ClusterIP", "NodePort", "LoadBalancer"
	Ports       []PortConfig      `json:"ports" yaml:"ports"`
	Ingress     *IngressConfig    `json:"ingress" yaml:"ingress"`
	DNS         *DNSConfig        `json:"dns" yaml:"dns"`
	TLS         *TLSConfig        `json:"tls" yaml:"tls"`
	Proxy       *ProxyConfig      `json:"proxy" yaml:"proxy"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

// PortConfig defines port configuration
type PortConfig struct {
	Name       string `json:"name" yaml:"name"`
	Port       int    `json:"port" yaml:"port"`
	TargetPort int    `json:"target_port" yaml:"target_port"`
	Protocol   string `json:"protocol" yaml:"protocol"`
	AgentID    string `json:"agent_id" yaml:"agent_id"`
}

// IngressConfig defines ingress configuration
type IngressConfig struct {
	Enabled     bool              `json:"enabled" yaml:"enabled"`
	ClassName   string            `json:"class_name" yaml:"class_name"`
	Hosts       []string          `json:"hosts" yaml:"hosts"`
	Rules       []IngressRule     `json:"rules" yaml:"rules"`
	TLS         []IngressTLS      `json:"tls" yaml:"tls"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

// IngressRule defines ingress rule
type IngressRule struct {
	Host    string        `json:"host" yaml:"host"`
	Paths   []IngressPath `json:"paths" yaml:"paths"`
	AgentID string        `json:"agent_id" yaml:"agent_id"`
}

// IngressPath defines ingress path
type IngressPath struct {
	Path        string `json:"path" yaml:"path"`
	PathType    string `json:"path_type" yaml:"path_type"`
	ServiceName string `json:"service_name" yaml:"service_name"`
	ServicePort int    `json:"service_port" yaml:"service_port"`
	AgentID     string `json:"agent_id" yaml:"agent_id"`
}

// IngressTLS defines TLS configuration for ingress
type IngressTLS struct {
	Hosts      []string `json:"hosts" yaml:"hosts"`
	SecretName string   `json:"secret_name" yaml:"secret_name"`
}

// DNSConfig defines DNS configuration
type DNSConfig struct {
	Policy      string   `json:"policy" yaml:"policy"`
	Nameservers []string `json:"nameservers" yaml:"nameservers"`
	Searches    []string `json:"searches" yaml:"searches"`
	Options     []string `json:"options" yaml:"options"`
}

// TLSConfig defines TLS configuration
type TLSConfig struct {
	Enabled            bool   `json:"enabled" yaml:"enabled"`
	CertFile           string `json:"cert_file" yaml:"cert_file"`
	KeyFile            string `json:"key_file" yaml:"key_file"`
	CAFile             string `json:"ca_file" yaml:"ca_file"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

// ProxyConfig defines proxy configuration
type ProxyConfig struct {
	Enabled      bool              `json:"enabled" yaml:"enabled"`
	Type         string            `json:"type" yaml:"type"` // "http", "socks5"
	URL          string            `json:"url" yaml:"url"`
	Headers      map[string]string `json:"headers" yaml:"headers"`
	Timeout      time.Duration     `json:"timeout" yaml:"timeout"`
	Retries      int               `json:"retries" yaml:"retries"`
	MaxIdleConns int               `json:"max_idle_conns" yaml:"max_idle_conns"`
}

// ScalingConfig defines scaling configuration
type ScalingConfig struct {
	Enabled             bool                 `json:"enabled" yaml:"enabled"`
	MinReplicas         int                  `json:"min_replicas" yaml:"min_replicas"`
	MaxReplicas         int                  `json:"max_replicas" yaml:"max_replicas"`
	TargetCPUPercent    int                  `json:"target_cpu_percent" yaml:"target_cpu_percent"`
	TargetMemoryPercent int                  `json:"target_memory_percent" yaml:"target_memory_percent"`
	ScaleUpCooldown     time.Duration        `json:"scale_up_cooldown" yaml:"scale_up_cooldown"`
	ScaleDownCooldown   time.Duration        `json:"scale_down_cooldown" yaml:"scale_down_cooldown"`
	CustomMetrics       []CustomMetricConfig `json:"custom_metrics" yaml:"custom_metrics"`
}

// CustomMetricConfig defines custom metric for scaling
type CustomMetricConfig struct {
	Name        string `json:"name" yaml:"name"`
	Type        string `json:"type" yaml:"type"`
	Query       string `json:"query" yaml:"query"`
	TargetValue string `json:"target_value" yaml:"target_value"`
	Resource    string `json:"resource" yaml:"resource"`
}

// HealthCheckConfig defines health check configuration
type HealthCheckConfig struct {
	Enabled             bool                          `json:"enabled" yaml:"enabled"`
	Path                string                        `json:"path" yaml:"path"`
	Port                int                           `json:"port" yaml:"port"`
	InitialDelaySeconds int                           `json:"initial_delay_seconds" yaml:"initial_delay_seconds"`
	PeriodSeconds       int                           `json:"period_seconds" yaml:"period_seconds"`
	TimeoutSeconds      int                           `json:"timeout_seconds" yaml:"timeout_seconds"`
	SuccessThreshold    int                           `json:"success_threshold" yaml:"success_threshold"`
	FailureThreshold    int                           `json:"failure_threshold" yaml:"failure_threshold"`
	HTTPHeaders         []HTTPHeader                  `json:"http_headers" yaml:"http_headers"`
	AgentSpecific       map[string]*HealthCheckConfig `json:"agent_specific" yaml:"agent_specific"`
}

// HTTPHeader defines HTTP header for health checks
type HTTPHeader struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// VolumeConfig defines volume configuration
type VolumeConfig struct {
	Name      string `json:"name" yaml:"name"`
	Type      string `json:"type" yaml:"type"` // "configmap", "secret", "emptydir", "persistentvolume"
	Source    string `json:"source" yaml:"source"`
	MountPath string `json:"mount_path" yaml:"mount_path"`
	ReadOnly  bool   `json:"read_only" yaml:"read_only"`
	AgentID   string `json:"agent_id" yaml:"agent_id"`
}

// SharedConfig defines shared configuration for all agents
type SharedConfig struct {
	Database     *DatabaseConfig               `json:"database" yaml:"database"`
	Cache        *CacheConfig                  `json:"cache" yaml:"cache"`
	Logging      *LoggingConfig                `json:"logging" yaml:"logging"`
	Monitoring   *MonitoringConfig             `json:"monitoring" yaml:"monitoring"`
	Security     *SecurityConfig               `json:"security" yaml:"security"`
	Environment  map[string]string             `json:"environment" yaml:"environment"`
	Secrets      map[string]string             `json:"secrets" yaml:"secrets"`
	LLMProviders map[string]*LLMProviderConfig `json:"llm_providers" yaml:"llm_providers"`
}

// DatabaseConfig defines database configuration
type DatabaseConfig struct {
	Type         string        `json:"type" yaml:"type"`
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	Database     string        `json:"database" yaml:"database"`
	Username     string        `json:"username" yaml:"username"`
	Password     string        `json:"password" yaml:"password"`
	SSLMode      string        `json:"ssl_mode" yaml:"ssl_mode"`
	MaxConns     int           `json:"max_conns" yaml:"max_conns"`
	MaxIdleConns int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxLifetime  time.Duration `json:"max_lifetime" yaml:"max_lifetime"`
}

// CacheConfig defines cache configuration
type CacheConfig struct {
	Type       string        `json:"type" yaml:"type"`
	Host       string        `json:"host" yaml:"host"`
	Port       int           `json:"port" yaml:"port"`
	Password   string        `json:"password" yaml:"password"`
	Database   int           `json:"database" yaml:"database"`
	TTL        time.Duration `json:"ttl" yaml:"ttl"`
	MaxRetries int           `json:"max_retries" yaml:"max_retries"`
}

// LoggingConfig defines logging configuration
type LoggingConfig struct {
	Level      string            `json:"level" yaml:"level"`
	Format     string            `json:"format" yaml:"format"`
	Output     []string          `json:"output" yaml:"output"`
	Structured bool              `json:"structured" yaml:"structured"`
	Fields     map[string]string `json:"fields" yaml:"fields"`
}

// MonitoringConfig defines monitoring configuration
type MonitoringConfig struct {
	Enabled  bool            `json:"enabled" yaml:"enabled"`
	Metrics  *MetricsConfig  `json:"metrics" yaml:"metrics"`
	Tracing  *TracingConfig  `json:"tracing" yaml:"tracing"`
	Alerting *AlertingConfig `json:"alerting" yaml:"alerting"`
}

// MetricsConfig defines metrics configuration
type MetricsConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Path      string `json:"path" yaml:"path"`
	Port      int    `json:"port" yaml:"port"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Subsystem string `json:"subsystem" yaml:"subsystem"`
}

// TracingConfig defines tracing configuration
type TracingConfig struct {
	Enabled     bool    `json:"enabled" yaml:"enabled"`
	Endpoint    string  `json:"endpoint" yaml:"endpoint"`
	ServiceName string  `json:"service_name" yaml:"service_name"`
	SampleRate  float64 `json:"sample_rate" yaml:"sample_rate"`
}

// AlertingConfig defines alerting configuration
type AlertingConfig struct {
	Enabled  bool            `json:"enabled" yaml:"enabled"`
	Webhooks []WebhookConfig `json:"webhooks" yaml:"webhooks"`
	Email    *EmailConfig    `json:"email" yaml:"email"`
	Slack    *SlackConfig    `json:"slack" yaml:"slack"`
	Rules    []AlertRule     `json:"rules" yaml:"rules"`
}

// WebhookConfig defines webhook configuration
type WebhookConfig struct {
	URL     string            `json:"url" yaml:"url"`
	Method  string            `json:"method" yaml:"method"`
	Headers map[string]string `json:"headers" yaml:"headers"`
	Timeout time.Duration     `json:"timeout" yaml:"timeout"`
}

// EmailConfig defines email configuration
type EmailConfig struct {
	SMTP     *SMTPConfig `json:"smtp" yaml:"smtp"`
	From     string      `json:"from" yaml:"from"`
	To       []string    `json:"to" yaml:"to"`
	Subject  string      `json:"subject" yaml:"subject"`
	Template string      `json:"template" yaml:"template"`
}

// SMTPConfig defines SMTP configuration
type SMTPConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	TLS      bool   `json:"tls" yaml:"tls"`
}

// SlackConfig defines Slack configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url" yaml:"webhook_url"`
	Channel    string `json:"channel" yaml:"channel"`
	Username   string `json:"username" yaml:"username"`
	IconEmoji  string `json:"icon_emoji" yaml:"icon_emoji"`
}

// AlertRule defines alert rule
type AlertRule struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Metric      string            `json:"metric" yaml:"metric"`
	Condition   string            `json:"condition" yaml:"condition"`
	Threshold   float64           `json:"threshold" yaml:"threshold"`
	Duration    time.Duration     `json:"duration" yaml:"duration"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

// SecurityConfig defines security configuration
type SecurityConfig struct {
	Authentication *AuthConfig       `json:"authentication" yaml:"authentication"`
	Authorization  *AuthzConfig      `json:"authorization" yaml:"authorization"`
	Encryption     *EncryptionConfig `json:"encryption" yaml:"encryption"`
	RateLimit      *RateLimitConfig  `json:"rate_limit" yaml:"rate_limit"`
	CORS           *CORSConfig       `json:"cors" yaml:"cors"`
	Headers        map[string]string `json:"headers" yaml:"headers"`
}

// AuthConfig defines authentication configuration
type AuthConfig struct {
	Type     string                 `json:"type" yaml:"type"` // "jwt", "oauth", "basic", "apikey"
	Config   map[string]interface{} `json:"config" yaml:"config"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Required bool                   `json:"required" yaml:"required"`
}

// AuthzConfig defines authorization configuration
type AuthzConfig struct {
	Type    string                 `json:"type" yaml:"type"` // "rbac", "acl", "policy"
	Config  map[string]interface{} `json:"config" yaml:"config"`
	Enabled bool                   `json:"enabled" yaml:"enabled"`
}

// EncryptionConfig defines encryption configuration
type EncryptionConfig struct {
	AtRest    *EncryptionAtRestConfig    `json:"at_rest" yaml:"at_rest"`
	InTransit *EncryptionInTransitConfig `json:"in_transit" yaml:"in_transit"`
}

// EncryptionAtRestConfig defines encryption at rest configuration
type EncryptionAtRestConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Algorithm string `json:"algorithm" yaml:"algorithm"`
	KeyID     string `json:"key_id" yaml:"key_id"`
	Provider  string `json:"provider" yaml:"provider"`
}

// EncryptionInTransitConfig defines encryption in transit configuration
type EncryptionInTransitConfig struct {
	Enabled  bool     `json:"enabled" yaml:"enabled"`
	MinTLS   string   `json:"min_tls" yaml:"min_tls"`
	Ciphers  []string `json:"ciphers" yaml:"ciphers"`
	CertFile string   `json:"cert_file" yaml:"cert_file"`
	KeyFile  string   `json:"key_file" yaml:"key_file"`
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Enabled    bool                  `json:"enabled" yaml:"enabled"`
	Global     *RateLimit            `json:"global" yaml:"global"`
	PerAgent   map[string]*RateLimit `json:"per_agent" yaml:"per_agent"`
	PerUser    *RateLimit            `json:"per_user" yaml:"per_user"`
	PerIP      *RateLimit            `json:"per_ip" yaml:"per_ip"`
	BurstLimit int                   `json:"burst_limit" yaml:"burst_limit"`
	WindowSize time.Duration         `json:"window_size" yaml:"window_size"`
}

// RateLimit defines rate limit configuration
type RateLimit struct {
	Requests  int           `json:"requests" yaml:"requests"`
	Period    time.Duration `json:"period" yaml:"period"`
	Burst     int           `json:"burst" yaml:"burst"`
	SkipPaths []string      `json:"skip_paths" yaml:"skip_paths"`
}

// CORSConfig defines CORS configuration
type CORSConfig struct {
	Enabled          bool     `json:"enabled" yaml:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers" yaml:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `json:"max_age" yaml:"max_age"`
}

// LLMProviderConfig defines LLM provider configuration
type LLMProviderConfig struct {
	Type       string                 `json:"type" yaml:"type"`
	APIKey     string                 `json:"api_key" yaml:"api_key"`
	Endpoint   string                 `json:"endpoint" yaml:"endpoint"`
	Model      string                 `json:"model" yaml:"model"`
	Config     map[string]interface{} `json:"config" yaml:"config"`
	Timeout    time.Duration          `json:"timeout" yaml:"timeout"`
	MaxRetries int                    `json:"max_retries" yaml:"max_retries"`
}

// ExtendedAgentConfig extends AgentConfig with additional deployment-specific fields
type ExtendedAgentConfig struct {
	*AgentConfig
	Path         string             `json:"path" yaml:"path"`
	Port         int                `json:"port" yaml:"port"`
	Host         string             `json:"host" yaml:"host"`
	Subdomain    string             `json:"subdomain" yaml:"subdomain"`
	Schema       *SchemaConfig      `json:"schema" yaml:"schema"`
	Middleware   []MiddlewareConfig `json:"middleware" yaml:"middleware"`
	Resources    *ResourceConfig    `json:"resources" yaml:"resources"`
	Scaling      *ScalingConfig     `json:"scaling" yaml:"scaling"`
	Environment  map[string]string  `json:"environment" yaml:"environment"`
	Secrets      map[string]string  `json:"secrets" yaml:"secrets"`
	ConfigMaps   map[string]string  `json:"config_maps" yaml:"config_maps"`
	Volumes      []VolumeConfig     `json:"volumes" yaml:"volumes"`
	Dependencies []string           `json:"dependencies" yaml:"dependencies"`
	Priority     int                `json:"priority" yaml:"priority"`
	Labels       map[string]string  `json:"labels" yaml:"labels"`
	Annotations  map[string]string  `json:"annotations" yaml:"annotations"`
	Disabled     bool               `json:"disabled" yaml:"disabled"`
}

// SchemaConfig defines input/output schema validation
type SchemaConfig struct {
	Input  *SchemaDefinition `json:"input" yaml:"input"`
	Output *SchemaDefinition `json:"output" yaml:"output"`
}

// SchemaDefinition defines schema definition
type SchemaDefinition struct {
	Type       string                         `json:"type" yaml:"type"`
	Properties map[string]*PropertyDefinition `json:"properties" yaml:"properties"`
	Required   []string                       `json:"required" yaml:"required"`
	MinLength  int                            `json:"min_length" yaml:"min_length"`
	MaxLength  int                            `json:"max_length" yaml:"max_length"`
	Pattern    string                         `json:"pattern" yaml:"pattern"`
	Format     string                         `json:"format" yaml:"format"`
	Enum       []interface{}                  `json:"enum" yaml:"enum"`
	Example    interface{}                    `json:"example" yaml:"example"`
}

// PropertyDefinition defines property definition
type PropertyDefinition struct {
	Type        string                         `json:"type" yaml:"type"`
	Description string                         `json:"description" yaml:"description"`
	Properties  map[string]*PropertyDefinition `json:"properties" yaml:"properties"`
	Items       *PropertyDefinition            `json:"items" yaml:"items"`
	Required    []string                       `json:"required" yaml:"required"`
	MinLength   int                            `json:"min_length" yaml:"min_length"`
	MaxLength   int                            `json:"max_length" yaml:"max_length"`
	Minimum     float64                        `json:"minimum" yaml:"minimum"`
	Maximum     float64                        `json:"maximum" yaml:"maximum"`
	Pattern     string                         `json:"pattern" yaml:"pattern"`
	Format      string                         `json:"format" yaml:"format"`
	Enum        []interface{}                  `json:"enum" yaml:"enum"`
	Default     interface{}                    `json:"default" yaml:"default"`
	Example     interface{}                    `json:"example" yaml:"example"`
}

// DefaultMultiAgentConfig returns default multi-agent configuration
func DefaultMultiAgentConfig() *MultiAgentConfig {
	return &MultiAgentConfig{
		Version:  "1.0.0",
		Agents:   make(map[string]*AgentConfig),
		Metadata: make(map[string]interface{}),
		Routing: &RoutingConfig{
			Type:       "path",
			Rules:      []RoutingRule{},
			Middleware: []MiddlewareConfig{},
		},
		Deployment: &DeploymentConfig{
			Type:        "docker",
			Environment: "development",
			Replicas:    1,
			Resources: &ResourceConfig{
				CPU:    "500m",
				Memory: "512Mi",
			},
			Networking: &NetworkingConfig{
				Type: "ClusterIP",
				Ports: []PortConfig{
					{
						Name:       "http",
						Port:       8080,
						TargetPort: 8080,
						Protocol:   "TCP",
					},
				},
			},
			HealthCheck: &HealthCheckConfig{
				Enabled:             true,
				Path:                "/health",
				Port:                8080,
				InitialDelaySeconds: 30,
				PeriodSeconds:       10,
				TimeoutSeconds:      5,
				SuccessThreshold:    1,
				FailureThreshold:    3,
				AgentSpecific:       make(map[string]*HealthCheckConfig),
			},
			Secrets:    make(map[string]string),
			ConfigMaps: make(map[string]string),
			Volumes:    []VolumeConfig{},
			Metadata:   make(map[string]interface{}),
		},
		Shared: &SharedConfig{
			Environment:  make(map[string]string),
			Secrets:      make(map[string]string),
			LLMProviders: make(map[string]*LLMProviderConfig),
			Logging: &LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     []string{"stdout"},
				Structured: true,
				Fields:     make(map[string]string),
			},
			Monitoring: &MonitoringConfig{
				Enabled: true,
				Metrics: &MetricsConfig{
					Enabled:   true,
					Path:      "/metrics",
					Port:      9090,
					Namespace: "golanggraph",
				},
				Tracing: &TracingConfig{
					Enabled:     false,
					SampleRate:  0.1,
					ServiceName: "golanggraph-multi-agent",
				},
			},
			Security: &SecurityConfig{
				CORS: &CORSConfig{
					Enabled:          true,
					AllowedOrigins:   []string{"*"},
					AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:   []string{"*"},
					AllowCredentials: false,
					MaxAge:           86400,
				},
				Headers: make(map[string]string),
			},
		},
	}
}

// Validate validates the multi-agent configuration
func (mac *MultiAgentConfig) Validate() error {
	if mac.Name == "" {
		return fmt.Errorf("multi-agent config name is required")
	}

	if len(mac.Agents) == 0 {
		return fmt.Errorf("at least one agent must be defined")
	}

	// Validate agent configs
	for agentID, agentConfig := range mac.Agents {
		if agentConfig.Name == "" {
			return fmt.Errorf("agent %s: name is required", agentID)
		}
		if agentConfig.Type == "" {
			return fmt.Errorf("agent %s: type is required", agentID)
		}
		if agentConfig.Model == "" {
			return fmt.Errorf("agent %s: model is required", agentID)
		}
		if agentConfig.Provider == "" {
			return fmt.Errorf("agent %s: provider is required", agentID)
		}
	}

	// Validate routing configuration
	if mac.Routing != nil {
		if err := mac.validateRouting(); err != nil {
			return fmt.Errorf("routing validation failed: %w", err)
		}
	}

	// Validate deployment configuration
	if mac.Deployment != nil {
		if err := mac.validateDeployment(); err != nil {
			return fmt.Errorf("deployment validation failed: %w", err)
		}
	}

	return nil
}

// validateRouting validates routing configuration
func (mac *MultiAgentConfig) validateRouting() error {
	if mac.Routing.DefaultAgent != "" {
		if _, exists := mac.Agents[mac.Routing.DefaultAgent]; !exists {
			return fmt.Errorf("default agent %s does not exist", mac.Routing.DefaultAgent)
		}
	}

	for _, rule := range mac.Routing.Rules {
		if rule.AgentID == "" {
			return fmt.Errorf("routing rule %s: agent_id is required", rule.ID)
		}
		if _, exists := mac.Agents[rule.AgentID]; !exists {
			return fmt.Errorf("routing rule %s: agent %s does not exist", rule.ID, rule.AgentID)
		}
	}

	return nil
}

// validateDeployment validates deployment configuration
func (mac *MultiAgentConfig) validateDeployment() error {
	if mac.Deployment.Type == "" {
		return fmt.Errorf("deployment type is required")
	}

	if mac.Deployment.Replicas < 1 {
		return fmt.Errorf("deployment replicas must be at least 1")
	}

	// Validate port assignments if networking is configured
	if mac.Deployment.Networking != nil {
		portMap := make(map[int]string)
		for _, port := range mac.Deployment.Networking.Ports {
			if existing, exists := portMap[port.Port]; exists {
				return fmt.Errorf("port %d is assigned to both %s and %s", port.Port, existing, port.Name)
			}
			portMap[port.Port] = port.Name
		}
	}

	return nil
}

// GetAgentByPath returns the agent ID for a given path
func (mac *MultiAgentConfig) GetAgentByPath(path string) (string, bool) {
	if mac.Routing == nil {
		return "", false
	}

	// Check routing rules
	for _, rule := range mac.Routing.Rules {
		if mac.matchesRule(rule, path) {
			return rule.AgentID, true
		}
	}

	// Return default agent if configured
	if mac.Routing.DefaultAgent != "" {
		return mac.Routing.DefaultAgent, true
	}

	return "", false
}

// matchesRule checks if a path matches a routing rule
func (mac *MultiAgentConfig) matchesRule(rule RoutingRule, path string) bool {
	switch rule.Pattern {
	case "prefix":
		return strings.HasPrefix(path, rule.Pattern)
	case "suffix":
		return strings.HasSuffix(path, rule.Pattern)
	case "exact":
		return path == rule.Pattern
	case "contains":
		return strings.Contains(path, rule.Pattern)
	default:
		// Default to prefix matching
		return strings.HasPrefix(path, rule.Pattern)
	}
}

// GetAgentPaths returns all paths for each agent
func (mac *MultiAgentConfig) GetAgentPaths() map[string][]string {
	paths := make(map[string][]string)

	if mac.Routing == nil {
		return paths
	}

	for _, rule := range mac.Routing.Rules {
		paths[rule.AgentID] = append(paths[rule.AgentID], rule.Pattern)
	}

	return paths
}

// GetAgentPort returns the port for a specific agent
func (mac *MultiAgentConfig) GetAgentPort(agentID string) int {
	if mac.Deployment == nil || mac.Deployment.Networking == nil {
		return 8080 // Default port
	}

	for _, port := range mac.Deployment.Networking.Ports {
		if port.AgentID == agentID {
			return port.Port
		}
	}

	return 8080 // Default port
}

// ListAgentIDs returns all agent IDs
func (mac *MultiAgentConfig) ListAgentIDs() []string {
	ids := make([]string, 0, len(mac.Agents))
	for id := range mac.Agents {
		ids = append(ids, id)
	}
	return ids
}

// GetEnabledAgents returns only enabled agents
func (mac *MultiAgentConfig) GetEnabledAgents() map[string]*AgentConfig {
	enabled := make(map[string]*AgentConfig)
	for id, config := range mac.Agents {
		// Check if agent is disabled (assuming we extend AgentConfig with a Disabled field)
		// For now, all agents are considered enabled
		enabled[id] = config
	}
	return enabled
}
