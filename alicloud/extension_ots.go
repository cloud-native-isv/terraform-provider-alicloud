package alicloud

import (
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore/search"
)

type PrimaryKeyTypeString string

const (
	IntegerType = PrimaryKeyTypeString("Integer")
	StringType  = PrimaryKeyTypeString("String")
	BinaryType  = PrimaryKeyTypeString("Binary")
)

type DefinedColumnTypeString string

const (
	DefinedColumnInteger = DefinedColumnTypeString("Integer")
	DefinedColumnString  = DefinedColumnTypeString("String")
	DefinedColumnBinary  = DefinedColumnTypeString("Binary")
	DefinedColumnDouble  = DefinedColumnTypeString("Double")
	DefinedColumnBoolean = DefinedColumnTypeString("Boolean")
)

type InstanceAccessedByType string

const (
	AnyNetwork   = InstanceAccessedByType("Any")
	VpcOnly      = InstanceAccessedByType("Vpc")
	VpcOrConsole = InstanceAccessedByType("ConsoleOrVpc")
)

type OtsInstanceType string

const (
	OtsCapacity        = OtsInstanceType("Capacity")
	OtsHighPerformance = OtsInstanceType("HighPerformance")
)

type OtsNetworkType string

const (
	VpcAccess      = OtsNetworkType("VPC")
	InternetAccess = OtsNetworkType("INTERNET")
	ClassicAccess  = OtsNetworkType("CLASSIC")
)

type OtsNetworkSource string

const (
	TrustProxyAccess = OtsNetworkSource("TRUST_PROXY")
)

func convertInstanceAccessedBy(accessed InstanceAccessedByType) string {
	switch accessed {
	case VpcOnly:
		return "VPC"
	case VpcOrConsole:
		return "VPC_CONSOLE"
	default:
		return "NORMAL"
	}
}

func convertInstanceAccessedByRevert(network string) InstanceAccessedByType {
	switch network {
	case "VPC":
		return VpcOnly
	case "VPC_CONSOLE":
		return VpcOrConsole
	default:
		return AnyNetwork
	}
}

func convertInstanceType(instanceType OtsInstanceType) string {
	switch instanceType {
	case OtsHighPerformance:
		return "SSD"
	default:
		return "HYBRID"
	}
}

func convertInstanceTypeRevert(instanceType string) OtsInstanceType {
	switch instanceType {
	case "SSD":
		return OtsHighPerformance
	default:
		return OtsCapacity
	}
}

func toInstanceOuterStatus(otsInstanceInnerStatus string) Status {
	switch otsInstanceInnerStatus {
	case "normal":
		return Running
	case "forbidden":
		return DisabledStatus
	case "deleting":
		return Deleting
	default:
		return Status(otsInstanceInnerStatus)
	}
}

func toInstanceInnerStatus(instanceOuterStatus Status) string {
	switch instanceOuterStatus {
	case Running:
		return "normal"
	case DisabledStatus:
		return "forbidden"
	case Deleting:
		return "deleting"
	default:
		return "INVALID"
	}
}

type TunnelTypeString string

const (
	BaseAndStreamTunnel = TunnelTypeString("BaseAndStream")
	BaseDataTunnel      = TunnelTypeString("BaseData")
	StreamTunnel        = TunnelTypeString("Stream")
)

type SseKeyTypeString string

const (
	SseKMSService = SseKeyTypeString("SSE_KMS_SERVICE")
	SseByOk       = SseKeyTypeString("SSE_BYOK")
)

type IndexTypeString string

const (
	Local  = IndexTypeString("Local")
	Global = IndexTypeString("Global")
)
const (
	SearchIndexTypeHolder = "Search"
)

type OtsSearchIndexSyncPhaseString string

const (
	Full = OtsSearchIndexSyncPhaseString("Full")
	Incr = OtsSearchIndexSyncPhaseString("Incr")
)

type SearchIndexFieldTypeString string

const (
	OtsSearchTypeLong     = SearchIndexFieldTypeString("Long")
	OtsSearchTypeDouble   = SearchIndexFieldTypeString("Double")
	OtsSearchTypeBoolean  = SearchIndexFieldTypeString("Boolean")
	OtsSearchTypeKeyword  = SearchIndexFieldTypeString("Keyword")
	OtsSearchTypeText     = SearchIndexFieldTypeString("Text")
	OtsSearchTypeDate     = SearchIndexFieldTypeString("Date")
	OtsSearchTypeGeoPoint = SearchIndexFieldTypeString("GeoPoint")
	OtsSearchTypeNested   = SearchIndexFieldTypeString("Nested")
)

type SearchIndexAnalyzerTypeString string

const (
	OtsSearchSingleWord = SearchIndexAnalyzerTypeString("SingleWord")
	OtsSearchSplit      = SearchIndexAnalyzerTypeString("Split")
	OtsSearchMinWord    = SearchIndexAnalyzerTypeString("MinWord")
	OtsSearchMaxWord    = SearchIndexAnalyzerTypeString("MaxWord")
	OtsSearchFuzzy      = SearchIndexAnalyzerTypeString("Fuzzy")
)

type SearchIndexOrderTypeString string

const (
	OtsSearchSortOrderAsc  = SearchIndexOrderTypeString("Asc")
	OtsSearchSortOrderDesc = SearchIndexOrderTypeString("Desc")
)

type SearchIndexSortModeString string

const (
	OtsSearchModeMin = SearchIndexSortModeString("Min")
	OtsSearchModeMax = SearchIndexSortModeString("Max")
	OtsSearchModeAvg = SearchIndexSortModeString("Avg")
)

type SearchIndexSortFieldTypeString string

const (
	OtsSearchPrimaryKeySort = SearchIndexSortFieldTypeString("PrimaryKeySort")
	OtsSearchFieldSort      = SearchIndexSortFieldTypeString("FieldSort")
)

type RestOtsInstanceInfo struct {
	InstanceStatus        string           `json:"InstanceStatus" xml:"InstanceStatus"`
	InstanceSpecification string           `json:"InstanceSpecification" xml:"InstanceSpecification"`
	Timestamp             string           `json:"Timestamp" xml:"Timestamp"`
	UserId                string           `json:"UserId" xml:"UserId"`
	ResourceGroupId       string           `json:"ResourceGroupId" xml:"ResourceGroupId"`
	InstanceName          string           `json:"InstanceName" xml:"InstanceName"`
	CreateTime            string           `json:"CreateTime" xml:"CreateTime"`
	Network               string           `json:"Network" xml:"Network"`
	NetworkTypeACL        []string         `json:"NetworkTypeACL" xml:"NetworkTypeACL"`
	NetworkSourceACL      []string         `json:"NetworkSourceACL" xml:"NetworkSourceACL"`
	Policy                string           `json:"Policy" xml:"Policy"`
	PolicyVersion         int              `json:"PolicyVersion" xml:"PolicyVersion"`
	InstanceDescription   string           `json:"InstanceDescription" xml:"InstanceDescription"`
	Quota                 RestOtsQuota     `json:"Quota" xml:"Quota"`
	Tags                  []RestOtsTagInfo `json:"Tags" xml:"Tags"`
}

type RestOtsQuota struct {
	TableQuota int `json:"TableQuota" xml:"TableQuota"`
}

type RestOtsTagInfo struct {
	Key   string `json:"Key" xml:"Key"`
	Value string `json:"Value" xml:"Value"`
}

// ConvertSecIndexType converts tablestore IndexType to IndexTypeString
func ConvertSecIndexType(indexType tablestore.IndexType) (IndexTypeString, error) {
	switch indexType {
	case tablestore.IT_LOCAL_INDEX:
		return Local, nil
	case tablestore.IT_GLOBAL_INDEX:
		return Global, nil
	default:
		return "", fmt.Errorf("unknown secondary index type: %v", indexType)
	}
}

// ConvertSecIndexTypeString converts IndexTypeString to tablestore IndexType
func ConvertSecIndexTypeString(indexType IndexTypeString) (tablestore.IndexType, error) {
	switch indexType {
	case Local:
		return tablestore.IT_LOCAL_INDEX, nil
	case Global:
		return tablestore.IT_GLOBAL_INDEX, nil
	default:
		return tablestore.IT_GLOBAL_INDEX, fmt.Errorf("unknown secondary index type string: %s", indexType)
	}
}

// ConvertDefinedColumnType converts tablestore DefinedColumnType to DefinedColumnTypeString
func ConvertDefinedColumnType(columnType tablestore.DefinedColumnType) (DefinedColumnTypeString, error) {
	switch columnType {
	case tablestore.DefinedColumn_INTEGER:
		return DefinedColumnInteger, nil
	case tablestore.DefinedColumn_STRING:
		return DefinedColumnString, nil
	case tablestore.DefinedColumn_BINARY:
		return DefinedColumnBinary, nil
	case tablestore.DefinedColumn_DOUBLE:
		return DefinedColumnDouble, nil
	case tablestore.DefinedColumn_BOOLEAN:
		return DefinedColumnBoolean, nil
	default:
		return "", fmt.Errorf("unknown defined column type: %v", columnType)
	}
}

// ConvertSearchIndexFieldTypeString converts SearchIndexFieldTypeString to tablestore FieldType
func ConvertSearchIndexFieldTypeString(fieldType SearchIndexFieldTypeString) (tablestore.FieldType, error) {
	switch fieldType {
	case OtsSearchTypeLong:
		return tablestore.FieldType_LONG, nil
	case OtsSearchTypeDouble:
		return tablestore.FieldType_DOUBLE, nil
	case OtsSearchTypeBoolean:
		return tablestore.FieldType_BOOLEAN, nil
	case OtsSearchTypeKeyword:
		return tablestore.FieldType_KEYWORD, nil
	case OtsSearchTypeText:
		return tablestore.FieldType_TEXT, nil
	case OtsSearchTypeDate:
		return tablestore.FieldType_DATE, nil
	case OtsSearchTypeGeoPoint:
		return tablestore.FieldType_GEO_POINT, nil
	case OtsSearchTypeNested:
		return tablestore.FieldType_NESTED, nil
	default:
		return tablestore.FieldType_KEYWORD, fmt.Errorf("unknown search index field type: %s", fieldType)
	}
}

// ConvertSearchIndexAnalyzerTypeString converts SearchIndexAnalyzerTypeString to tablestore Analyzer
func ConvertSearchIndexAnalyzerTypeString(analyzer SearchIndexAnalyzerTypeString) (tablestore.Analyzer, error) {
	switch analyzer {
	case OtsSearchSingleWord:
		return tablestore.Analyzer_SingleWord, nil
	case OtsSearchSplit:
		return tablestore.Analyzer_Split, nil
	case OtsSearchMinWord:
		return tablestore.Analyzer_MinWord, nil
	case OtsSearchMaxWord:
		return tablestore.Analyzer_MaxWord, nil
	case OtsSearchFuzzy:
		return tablestore.Analyzer_Fuzzy, nil
	default:
		return tablestore.Analyzer_SingleWord, fmt.Errorf("unknown search index analyzer type: %s", analyzer)
	}
}

// ConvertSearchIndexSortFieldTypeString converts SearchIndexSortFieldTypeString to search Sorter
func ConvertSearchIndexSortFieldTypeString(sortType SearchIndexSortFieldTypeString) (search.Sorter, error) {
	switch sortType {
	case OtsSearchPrimaryKeySort:
		return &search.PrimaryKeySort{}, nil
	case OtsSearchFieldSort:
		return &search.FieldSort{}, nil
	default:
		return &search.PrimaryKeySort{}, fmt.Errorf("unknown search index sort field type: %s", sortType)
	}
}

// ConvertSearchIndexOrderTypeString converts SearchIndexOrderTypeString to search SortOrder
func ConvertSearchIndexOrderTypeString(order SearchIndexOrderTypeString) (search.SortOrder, error) {
	switch order {
	case OtsSearchSortOrderAsc:
		return search.SortOrder_ASC, nil
	case OtsSearchSortOrderDesc:
		return search.SortOrder_DESC, nil
	default:
		return search.SortOrder_ASC, fmt.Errorf("unknown search index order type: %s", order)
	}
}

// ConvertSearchIndexSortModeString converts SearchIndexSortModeString to search SortMode
func ConvertSearchIndexSortModeString(mode SearchIndexSortModeString) (search.SortMode, error) {
	switch mode {
	case OtsSearchModeMin:
		return search.SortMode_Min, nil
	case OtsSearchModeMax:
		return search.SortMode_Max, nil
	case OtsSearchModeAvg:
		return search.SortMode_Avg, nil
	default:
		return search.SortMode_Min, fmt.Errorf("unknown search index sort mode: %s", mode)
	}
}

// ConvertSearchIndexSyncPhase converts tablestore SyncPhase to OtsSearchIndexSyncPhaseString
func ConvertSearchIndexSyncPhase(phase tablestore.SyncPhase) (OtsSearchIndexSyncPhaseString, error) {
	switch phase {
	case tablestore.SyncPhase_FULL:
		return Full, nil
	case tablestore.SyncPhase_INCR:
		return Incr, nil
	default:
		return Full, fmt.Errorf("unknown search index sync phase: %v", phase)
	}
}

// ID generates an ID string from components
func ID(parts ...string) string {
	return strings.Join(parts, ":")
}
