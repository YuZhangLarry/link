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
	ChunkingConfig        *string    `json:"chunking_config,omitempty" gorm:"type:json"`
	ImageProcessingConfig *string    `json:"image_processing_config,omitempty" gorm:"type:json"`
	SummaryModelID        string     `json:"summary_model_id" gorm:"type:varchar(64)"`
	RerankModelID         string     `json:"rerank_model_id" gorm:"type:varchar(64)"`
	CosConfig             *string    `json:"cos_config,omitempty" gorm:"type:json"`
	VLMConfig             *string    `json:"vlm_config,omitempty" gorm:"type:json"`
	ExtractConfig         *string    `json:"extract_config,omitempty" gorm:"type:json"`
	Status                int8       `json:"status" gorm:"type:tinyint;default:1;index:idx_status"`
	IsPublic              bool       `json:"is_public" gorm:"default:false"`
	DocumentCount         int        `json:"document_count" gorm:"default:0"`
	ChunkCount            int        `json:"chunk_count" gorm:"default:0"`
	StorageSize           int64      `json:"storage_size" gorm:"default:0"`
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
	StorageSize      int64      `json:"storage_size" gorm:"default:0"`
	ChunkCount       int        `json:"chunk_count" gorm:"default:0"`
	Metadata         *string    `json:"metadata,omitempty" gorm:"type:json"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	ErrorMessage     string     `json:"error_message" gorm:"type:text"`
}

func (Knowledge) TableName() string {
	return "knowledge"
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
