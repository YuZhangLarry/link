package types

import "time"

// KnowledgeBase 知识库实体
type KnowledgeBase struct {
	ID                    string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TenantID              int64      `json:"tenant_id" gorm:"not null;index:idx_tenant_id"`
	UserID                int64      `json:"user_id" gorm:"not null;index:idx_user_id"`
	Name                  string     `json:"name" gorm:"type:varchar(100);not null"`
	Description           string     `json:"description" gorm:"type:text"`
	Avatar                string     `json:"avatar" gorm:"type:varchar(500)"`
	EmbeddingModelID      string     `json:"embedding_model_id" gorm:"type:varchar(64)"`
	ChunkingConfig        string     `gorm:"column:chunking_config;type:json;default:{}"`
	ImageProcessingConfig string     `gorm:"column:image_processing_config;type:json;default:{}"`
	SummaryModelID        string     `json:"summary_model_id" gorm:"type:varchar(64)"`
	RerankModelID         string     `json:"rerank_model_id" gorm:"type:varchar(64)"`
	CosConfig             string     `gorm:"column:cos_config;type:json;default:{}"`
	VLMConfig             string     `gorm:"column:vlm_config;type:json;default:{}"`
	ExtractConfig         string     `gorm:"column:extract_config;type:json;default:{}"`
	Status                int8       `json:"status" gorm:"type:tinyint;default:1;index:idx_status"`
	IsPublic              bool       `json:"is_public" gorm:"default:false"`
	CreatedAt             time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt             *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}

// Knowledge 知识条目实体
type Knowledge struct {
	ID               string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TenantID         int64      `json:"tenant_id" gorm:"not null;index:idx_tenant_kb,priority:1"`
	KBID             string     `json:"kb_id" gorm:"not null;type:varchar(36);index:idx_tenant_kb,priority:2"`
	UserID           int64      `json:"user_id" gorm:"not null;index:idx_user_id"`
	Type             string     `json:"type" gorm:"type:varchar(20);not null"`
	Title            string     `json:"title" gorm:"type:varchar(200);not null"`
	Description      string     `json:"description" gorm:"type:text"`
	Source           string     `json:"source" gorm:"type:varchar(20)"`
	ParseStatus      string     `json:"parse_status" gorm:"type:varchar(20);default:'unprocessed';index:idx_parse_status"`
	EnableStatus     string     `json:"enable_status" gorm:"type:varchar(20);default:'enabled';index:idx_enable_status"`
	TagID            int64      `json:"tag_id" gorm:"default:0"` // 关联的标签ID
	EmbeddingModelID string     `json:"embedding_model_id" gorm:"type:varchar(64)"`
	FileName         string     `json:"file_name" gorm:"type:varchar(255)"`
	FileType         string     `json:"file_type" gorm:"type:varchar(50)"`
	FileSize         int64      `json:"file_size" gorm:"default:0"`
	FilePath         string     `json:"file_path" gorm:"type:varchar(500)"`
	FileHash         string     `json:"file_hash" gorm:"type:varchar(64);index:idx_file_hash"`
	Metadata         *string    `json:"metadata,omitempty" gorm:"type:json"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	ErrorMessage     string     `json:"error_message" gorm:"type:text"`
}

func (Knowledge) TableName() string {
	return "knowledges"
}

// Chunk 文档分块实体
type Chunk struct {
	ID             string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TenantID       int64      `json:"tenant_id" gorm:"not null;index:idx_tenant_kb,priority:1"`
	KBID           string     `json:"kb_id" gorm:"not null;type:varchar(36);index:idx_tenant_kb,priority:2;index:idx_kb_id"`
	KnowledgeID    string     `json:"knowledge_id" gorm:"not null;type:varchar(36);index:idx_knowledge_id"`
	Content        string     `json:"content" gorm:"type:text;not null"`
	ChunkIndex     int        `json:"chunk_index" gorm:"not null;index:idx_chunk_index"`
	IsEnabled      bool       `json:"is_enabled" gorm:"default:true;index:idx_enabled"`
	StartAt        int        `json:"start_at" gorm:"default:0"`
	EndAt          int        `json:"end_at" gorm:"default:0"`
	PreChunkID     *string    `json:"pre_chunk_id,omitempty" gorm:"type:varchar(36)"`
	NextChunkID    *string    `json:"next_chunk_id,omitempty" gorm:"type:varchar(36)"`
	ChunkType      string     `json:"chunk_type" gorm:"type:varchar(20);default:'text'"`
	ParentChunkID  *string    `json:"parent_chunk_id,omitempty" gorm:"type:varchar(36)"`
	ImageInfo      *string    `json:"image_info,omitempty" gorm:"type:json"`
	RelationChunks *string    `json:"relation_chunks,omitempty" gorm:"type:json"`
	EmbeddingID    string     `json:"embedding_id" gorm:"type:varchar(64)"`
	TokenCount     int        `json:"token_count" gorm:"default:0"`
	TagID          int64      `json:"tag_id" gorm:"default:0"` // 关联的标签ID
	Metadata       *string    `json:"metadata,omitempty" gorm:"type:json"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Chunk) TableName() string {
	return "chunks"
}

// KBSetting 知识库设置实体
type KBSetting struct {
	ID                  int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	KBID                string    `json:"kb_id" gorm:"not null;type:varchar(36);uniqueIndex"`
	RetrievalMode       string    `json:"retrieval_mode" gorm:"type:varchar(20);default:'vector'"`
	SimilarityThreshold float64   `json:"similarity_threshold" gorm:"default:0.7"`
	TopK                int       `json:"top_k" gorm:"default:5"`
	RerankEnabled       bool      `json:"rerank_enabled" gorm:"default:false"`
	GraphEnabled        bool      `json:"graph_enabled" gorm:"default:false"`
	SettingsJSON        *string   `json:"settings_json,omitempty" gorm:"type:json"`
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (KBSetting) TableName() string {
	return "kb_settings"
}

// CreateKnowledgeBaseRequest 创建知识库请求
type CreateKnowledgeBaseRequest struct {
	Name             string `json:"name" binding:"required,min=1,max=100"`
	Description      string `json:"description" binding:"max=500"`
	Avatar           string `json:"avatar"`
	EmbeddingModelID string `json:"embedding_model_id"`
	IsPublic         bool   `json:"is_public"`
	ChunkSize        int    `json:"chunk_size"`
	ChunkOverlap     int    `json:"chunk_overlap"`
}

// UpdateKnowledgeBaseRequest 更新知识库请求
type UpdateKnowledgeBaseRequest struct {
	Name             *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description      *string `json:"description" binding:"omitempty,max=500"`
	Avatar           *string `json:"avatar"`
	EmbeddingModelID *string `json:"embedding_model_id"`
	SummaryModelID   *string `json:"summary_model_id"`
	RerankModelID    *string `json:"rerank_model_id"`
	IsPublic         *bool   `json:"is_public"`
	Status           *int8   `json:"status" binding:"omitempty,oneof=0 1"`
	ChunkingConfig   *string `json:"chunking_config"`
}

// KnowledgeBaseResponse 知识库响应
type KnowledgeBaseResponse struct {
	ID            string    `json:"id"`
	TenantID      int64     `json:"tenant_id"`
	UserID        int64     `json:"user_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Avatar        string    `json:"avatar"`
	DocumentCount int       `json:"document_count"`
	ChunkCount    int       `json:"chunk_count"`
	StorageSize   int64     `json:"storage_size"`
	Status        int8      `json:"status"`
	IsPublic      bool      `json:"is_public"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateKnowledgeRequest 创建知识条目请求
type CreateKnowledgeRequest struct {
	KBID        string `json:"kb_id" binding:"required"`
	Title       string `json:"title" binding:"required,max=200"`
	Description string `json:"description" binding:"max=1000"`
	Type        string `json:"type" binding:"required,oneof=document file url"`
	Source      string `json:"source" binding:"omitempty,oneof=upload crawler api"`
	FileName    string `json:"file_name"`
	FileType    string `json:"file_type"`
	FilePath    string `json:"file_path"`
	FileSize    int64  `json:"file_size"`
	Metadata    string `json:"metadata"`
}

// UpdateKnowledgeRequest 更新知识条目请求
type UpdateKnowledgeRequest struct {
	Title        *string `json:"title" binding:"omitempty,max=200"`
	Description  *string `json:"description" binding:"omitempty,max=1000"`
	EnableStatus *string `json:"enable_status" binding:"omitempty,oneof=enabled disabled"`
}

// KnowledgeResponse 知识条目响应
type KnowledgeResponse struct {
	ID           string     `json:"id"`
	TenantID     int64      `json:"tenant_id"`
	KBID         string     `json:"kb_id"`
	UserID       int64      `json:"user_id"`
	Type         string     `json:"type"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Source       string     `json:"source"`
	ParseStatus  string     `json:"parse_status"`
	EnableStatus string     `json:"enable_status"`
	FileName     string     `json:"file_name"`
	FileType     string     `json:"file_type"`
	FileSize     int64      `json:"file_size"`
	ChunkCount   int        `json:"chunk_count"`
	StorageSize  int64      `json:"storage_size"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
}

// ChunkResponse 分块响应
type ChunkResponse struct {
	ID          string    `json:"id"`
	KnowledgeID string    `json:"knowledge_id"`
	Content     string    `json:"content"`
	ChunkIndex  int       `json:"chunk_index"`
	IsEnabled   bool      `json:"is_enabled"`
	StartAt     int       `json:"start_at"`
	EndAt       int       `json:"end_at"`
	ChunkType   string    `json:"chunk_type"`
	TokenCount  int       `json:"token_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// KBSettingResponse 知识库设置响应
type KBSettingResponse struct {
	ID                  int64     `json:"id"`
	KBID                string    `json:"kb_id"`
	RetrievalMode       string    `json:"retrieval_mode"`
	SimilarityThreshold float64   `json:"similarity_threshold"`
	TopK                int       `json:"top_k"`
	RerankEnabled       bool      `json:"rerank_enabled"`
	GraphEnabled        bool      `json:"graph_enabled"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// KnowledgeListQuery 知识条目查询参数
type KnowledgeListQuery struct {
	KBID         string `form:"kb_id"`
	Type         string `form:"type"`
	ParseStatus  string `form:"parse_status"`
	EnableStatus string `form:"enable_status"`
	Page         int    `form:"page" binding:"min=1"`
	PageSize     int    `form:"page_size" binding:"min=1,max=100"`
}

// ChunkListQuery 分块查询参数
type ChunkListQuery struct {
	KBID        string `form:"kb_id"`
	KnowledgeID string `form:"knowledge_id"`
	IsEnabled   *bool  `form:"is_enabled"`
	Page        int    `form:"page" binding:"min=1"`
	PageSize    int    `form:"page_size" binding:"min=1,max=100"`
}

// ========================================
// 知识标签相关类型
// ========================================

// Tag 知识标签实体
type Tag struct {
	ID              int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID        string     `json:"tenant_id" gorm:"not null;type:varchar(36);index:idx_tenant_kb,priority:1"`
	KnowledgeBaseID int64      `json:"knowledge_base_id" gorm:"not null;index:idx_tenant_kb,priority:2"`
	Name            string     `json:"name" gorm:"type:varchar(255);not null;index:idx_name"`
	Color           string     `json:"color,omitempty" gorm:"type:varchar(7)"`
	SortOrder       int        `json:"sort_order" gorm:"default:0;index:idx_sort_order"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Tag) TableName() string {
	return "knowledge_tags"
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=255"`
	Color     string `json:"color" binding:"omitempty,hexcolor,len=7"`
	SortOrder int    `json:"sort_order"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name      *string `json:"name" binding:"omitempty,min=1,max=255"`
	Color     *string `json:"color" binding:"omitempty,hexcolor,len=7"`
	SortOrder *int    `json:"sort_order"`
}

// TagResponse 标签响应
type TagResponse struct {
	ID              int64     `json:"id"`
	TenantID        string    `json:"tenant_id"`
	KnowledgeBaseID int64     `json:"knowledge_base_id"`
	Name            string    `json:"name"`
	Color           string    `json:"color"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TagListQuery 标签查询参数
type TagListQuery struct {
	Name     string `form:"name"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
}

// ========================================
// 图谱关系类型相关定义
// ========================================

// RelationType 关系类型
type RelationType string

const (
	RelationTypeContains RelationType = "contains" // 包含
	RelationTypeRelates  RelationType = "relates"  // 关联
	RelationTypeDepends  RelationType = "depends"  // 依赖
	RelationTypeBelongs  RelationType = "belongs"  // 属于
	RelationTypeOwns     RelationType = "owns"     // 拥有
	RelationTypeAuthor   RelationType = "author"   // 作者
	RelationTypeAlias    RelationType = "alias"    // 别名
	RelationTypeOther    RelationType = "other"    // 其他
)

// RelationTypes 所有关系类型
var RelationTypes = []RelationType{
	RelationTypeContains,
	RelationTypeRelates,
	RelationTypeDepends,
	RelationTypeBelongs,
	RelationTypeOwns,
	RelationTypeAuthor,
	RelationTypeAlias,
	RelationTypeOther,
}

// IsValidRelationType 检查关系类型是否有效
func IsValidRelationType(t string) bool {
	switch RelationType(t) {
	case RelationTypeContains, RelationTypeRelates, RelationTypeDepends,
		RelationTypeBelongs, RelationTypeOwns, RelationTypeAuthor,
		RelationTypeAlias, RelationTypeOther:
		return true
	default:
		return false
	}
}

// RelationTypeLabel 获取关系类型的中文标签
func RelationTypeLabel(t string) string {
	labels := map[string]string{
		"contains": "包含",
		"relates":  "关联",
		"depends":  "依赖",
		"belongs":  "属于",
		"owns":     "拥有",
		"author":   "作者",
		"alias":    "别名",
		"other":    "其他",
	}
	if label, ok := labels[t]; ok {
		return label
	}
	return t
}

// RelationTypeOptions 获取关系类型选项（用于 API 响应）
func RelationTypeOptions() []RelationTypeOption {
	return []RelationTypeOption{
		{Value: "contains", Label: "包含"},
		{Value: "relates", Label: "关联"},
		{Value: "depends", Label: "依赖"},
		{Value: "belongs", Label: "属于"},
		{Value: "owns", Label: "拥有"},
		{Value: "author", Label: "作者"},
		{Value: "alias", Label: "别名"},
		{Value: "other", Label: "其他"},
	}
}

// RelationTypeOption 关系类型选项（用于 API 响应）
type RelationTypeOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ========================================
// 图谱节点和关系相关类型定义
// ========================================

// NameSpace 命名空间
type NameSpace struct {
	TenantID  string // 租户ID (对应 knowledge.tenant_id)
	KBID      string // 知识库ID (对应 knowledge.kb_id)
	Knowledge string // 知识条目ID (对应 knowledge.id)
	Type      string // 知识类型 (如 "document", "faq", "manual" 等)
}

// GraphData 图数据结构
type GraphData struct {
	Node     []*GraphNode     // 节点列表
	Relation []*GraphRelation // 关系列表
}

// GraphNode 图节点
type GraphNode struct {
	ID         string   `json:"id"`          // 节点唯一标识（UUID）
	Name       string   `json:"name"`        // 节点名称（实体名称）
	EntityType string   `json:"entity_type"` // 实体类型
	Attributes []string `json:"attributes"`  // 节点属性列表
	Chunks     []string `json:"chunks"`      // 关联的分块ID列表
}

// GraphRelation 图关系
type GraphRelation struct {
	ID             string   `json:"id"`              // 关系唯一标识（UUID）
	ChunkIDs       []string `json:"chunk_ids"`       // 记录该关系在哪些文档块中被识别到
	CombinedDegree int      `json:"combined_degree"` // 源实体和目标实体的度数之和
	Weight         float64  `json:"weight"`          // 关系强度权重，范围1-10，由PMI和Strength组合计算
	Source         string   `json:"source"`          // 关系起点的实体标题
	Target         string   `json:"target"`          // 关系终点的实体标题
	Type           string   `json:"type"`            // 关系类型
	Description    string   `json:"description"`     // 关系的语义描述（如"depends on", "contains"）
	Strength       float64  `json:"strength"`        // LLM提取的关系强度评分（1-10）
}
