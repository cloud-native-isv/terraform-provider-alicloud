package alicloud

// SlsLogtailPipelineConfig represents the domain model for Logtail Pipeline Config
type SlsLogtailPipelineConfig struct {
	Project        string
	Name           string
	Inputs         []SlsLogtailPipelineConfigPlugin
	Processors     []SlsLogtailPipelineConfigPlugin
	Flushers       []SlsLogtailPipelineConfigPlugin
	Aggregators    []SlsLogtailPipelineConfigPlugin
	GlobalJson     string
	TaskJson       string
	LogSample      string
	CreateTime     int64
	LastModifyTime int64
}

type SlsLogtailPipelineConfigPlugin struct {
	Type       string
	ConfigJson string
}
