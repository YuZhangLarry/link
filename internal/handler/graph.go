package handler

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"link/internal/application/service"
	"link/internal/middleware"
	"link/internal/types"
)

// GraphHandler 图谱处理器
type GraphHandler struct {
	graphService *service.GraphService
}

// NewGraphHandler 创建图谱处理器
func NewGraphHandler(graphService *service.GraphService) *GraphHandler {
	return &GraphHandler{
		graphService: graphService,
	}
}

// GetGraph 获取知识库图谱
// GET /api/v1/knowledge-bases/:id/graph
func (h *GraphHandler) GetGraph(c *gin.Context) {
	kbID := c.Param("id")

	// 使用 middleware 辅助函数获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 构建命名空间（使用 KBID 标签查询，不过滤 knowledge_id）
	namespace := types.NameSpace{
		TenantID: strconv.FormatInt(tenantID, 10),
		KBID:     kbID,
		// 不设置 Knowledge，让 repository 使用标签查询
	}

	// 获取图谱数据
	graphData, err := h.graphService.GetGraph(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    graphData,
	})
}

// SearchNode 搜索节点
// POST /api/v1/knowledge-bases/:id/graph/search
func (h *GraphHandler) SearchNode(c *gin.Context) {
	kbID := c.Param("id")

	// 使用 middleware 辅助函数获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 解析请求体
	var req struct {
		Nodes []string `json:"nodes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	if len(req.Nodes) == 0 {
		c.JSON(400, gin.H{"error": "nodes is required"})
		return
	}

	// 构建命名空间（使用 KBID 标签查询）
	namespace := types.NameSpace{
		TenantID: strconv.FormatInt(tenantID, 10),
		KBID:     kbID,
		// 不设置 Knowledge，让 repository 使用标签查询
	}

	// 搜索节点
	graphData, err := h.graphService.SearchNode(c.Request.Context(), namespace, req.Nodes)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    graphData,
	})
}

// GetNodeDetail 获取节点详情
// GET /api/v1/knowledge-bases/:id/graph/nodes/:nodeId
func (h *GraphHandler) GetNodeDetail(c *gin.Context) {
	kbID := c.Param("id")
	nodeTitle := c.Param("nodeId")

	// 使用 middleware 辅助函数获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 构建命名空间
	namespace := types.NameSpace{
		TenantID: strconv.FormatInt(tenantID, 10),
		KBID:     kbID,
	}

	// 搜索节点
	graphData, err := h.graphService.SearchNode(c.Request.Context(), namespace, []string{nodeTitle})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if len(graphData.Node) == 0 {
		c.JSON(404, gin.H{"error": "node not found"})
		return
	}

	// 获取节点及其相关关系
	node := graphData.Node[0]
	relations := graphData.Relation

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"node":      node,
			"relations": relations,
		},
	})
}

// AddNode 添加节点
// POST /api/v1/knowledge-bases/:id/graph/nodes
func (h *GraphHandler) AddNode(c *gin.Context) {
	kbID := c.Param("id")

	log.Printf("[Handler] AddNode START: kb_id=%s", kbID)

	// 使用 middleware 辅助函数获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		log.Printf("[Handler] AddNode ERROR: unauthorized, missing tenant_id")
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 解析请求体
	var req struct {
		Name       string   `json:"name" binding:"required"`
		EntityType string   `json:"entity_type"`
		Attributes []string `json:"attributes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Handler] AddNode ERROR: invalid request body: %v", err)
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	// 创建节点
	node := &types.GraphNode{
		ID:         uuid.New().String(),
		Name:       req.Name,
		EntityType: req.EntityType,
		Attributes: req.Attributes,
		Chunks:     []string{},
	}

	log.Printf("[Handler] AddNode Created node: ID=%s, Name=%q, EntityType=%q",
		node.ID, node.Name, node.EntityType)

	// 构建命名空间
	namespace := types.NameSpace{
		TenantID: strconv.FormatInt(tenantID, 10),
		KBID:     kbID,
	}

	log.Printf("[Handler] AddNode Namespace: TenantID=%s, KBID=%s", namespace.TenantID, namespace.KBID)
	log.Printf("[Handler] AddNode Calling Service.AddNode")

	// 添加节点（使用专门的 AddNode 方法）
	if err := h.graphService.AddNode(c.Request.Context(), namespace, node); err != nil {
		log.Printf("[Handler] AddNode ERROR from Service: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Handler] AddNode SUCCESS: node added with ID=%s, Name=%q", node.ID, node.Name)

	c.JSON(200, gin.H{
		"message": "success",
		"data":    node,
	})
}

// AddRelation 添加关系
// POST /api/v1/knowledge-bases/:id/graph/relations
func (h *GraphHandler) AddRelation(c *gin.Context) {
	kbID := c.Param("id")

	log.Printf("[Handler] AddRelation START: kb_id=%s", kbID)

	// 使用 middleware 辅助函数获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		log.Printf("[Handler] AddRelation ERROR: unauthorized, missing tenant_id")
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 解析请求体
	var req struct {
		Source   string  `json:"source" binding:"required"`
		Target   string  `json:"target" binding:"required"`
		Type     string  `json:"type" binding:"required"`
		Strength float64 `json:"strength"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Handler] AddRelation ERROR: invalid request body: %v", err)
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	log.Printf("[Handler] AddRelation REQUEST: source=%q, target=%q, type=%q, strength=%f",
		req.Source, req.Target, req.Type, req.Strength)

	// 创建关系
	relation := &types.GraphRelation{
		ID:       uuid.New().String(),
		ChunkIDs: []string{},
	}
	relation.Source = req.Source
	relation.Target = req.Target
	relation.Type = req.Type
	relation.Description = req.Type
	if req.Strength > 0 {
		relation.Strength = req.Strength
	} else {
		relation.Strength = 5.0
	}
	relation.Weight = relation.Strength

	log.Printf("[Handler] AddRelation Created relation: ID=%s, Source=%q, Target=%q, Type=%q, Strength=%f, Weight=%f",
		relation.ID, relation.Source, relation.Target, relation.Type, relation.Strength, relation.Weight)

	// 构建命名空间
	namespace := types.NameSpace{
		TenantID: strconv.FormatInt(tenantID, 10),
		KBID:     kbID,
	}

	log.Printf("[Handler] AddRelation Namespace: TenantID=%s, KBID=%s", namespace.TenantID, namespace.KBID)
	log.Printf("[Handler] AddRelation Calling Service.AddRelation")

	// 添加关系（使用专门的 AddRelation 方法）
	resultRelation, err := h.graphService.AddRelation(c.Request.Context(), namespace, relation)
	if err != nil {
		log.Printf("[Handler] AddRelation ERROR from Service: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if resultRelation == nil {
		log.Printf("[Handler] AddRelation WARNING: Service returned nil")
		c.JSON(500, gin.H{"error": "failed to add relation"})
		return
	}

	log.Printf("[Handler] AddRelation SUCCESS: relation added with ID=%s, Source=%q, Target=%q, Type=%q",
		resultRelation.ID, resultRelation.Source, resultRelation.Target, resultRelation.Type)

	c.JSON(200, gin.H{
		"message": "success",
		"data":    resultRelation,
	})
}

// UpdateNode 更新节点属性
// PUT /api/v1/knowledge-bases/:id/graph/nodes/:nodeId
func (h *GraphHandler) UpdateNode(c *gin.Context) {
	kbID := c.Param("id")
	nodeID := c.Param("nodeId")

	log.Printf("[Handler] UpdateNode START: kb_id=%s, node_id=%s", kbID, nodeID)

	// 使用 middleware 辅助函数获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		log.Printf("[Handler] UpdateNode ERROR: unauthorized, missing tenant_id")
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 解析请求体
	var req struct {
		Name       string   `json:"name" binding:"required"`
		EntityType string   `json:"entity_type"`
		Attributes []string `json:"attributes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Handler] UpdateNode ERROR: invalid request body: %v", err)
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	log.Printf("[Handler] UpdateNode REQUEST: name=%q, entity_type=%q", req.Name, req.EntityType)

	// 构建节点对象
	node := &types.GraphNode{
		ID:         nodeID,
		Name:       req.Name,
		EntityType: req.EntityType,
		Attributes: req.Attributes,
	}

	log.Printf("[Handler] UpdateNode Calling Service with node.ID=%s, Name=%q", node.ID, node.Name)

	// 构建命名空间
	namespace := types.NameSpace{
		TenantID: strconv.FormatInt(tenantID, 10),
		KBID:     kbID,
		// 不设置 Knowledge，让 repository 使用标签查询
	}

	// 更新节点
	if err := h.graphService.UpdateNode(c.Request.Context(), namespace, node); err != nil {
		log.Printf("[Handler] UpdateNode ERROR from Service: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Handler] UpdateNode SUCCESS: node updated with ID=%s, Name=%q", node.ID, node.Name)

	c.JSON(200, gin.H{
		"message": "success",
		"data":    node,
	})
}

// UpdateRelation 更新关系属性
// PUT /api/v1/knowledge-bases/:id/graph/relations/:relationId
func (h *GraphHandler) UpdateRelation(c *gin.Context) {
	kbID := c.Param("id")
	relationID := c.Param("relationId")

	log.Printf("[Handler] UpdateRelation START: kb_id=%s, relation_id=%s", kbID, relationID)

	// 使用 middleware 辅助函数获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		log.Printf("[Handler] UpdateRelation ERROR: unauthorized, missing tenant_id")
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 解析请求体
	var req struct {
		Type        string  `json:"type" binding:"required"`
		Description string  `json:"description"`
		Strength    float64 `json:"strength"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Handler] UpdateRelation ERROR: invalid request body: %v", err)
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	log.Printf("[Handler] UpdateRelation REQUEST: type=%q, description=%q, strength=%f", req.Type, req.Description, req.Strength)

	// 构建关系对象
	relation := &types.GraphRelation{
		ID:          relationID,
		Type:        req.Type,
		Description: req.Description,
		Strength:    req.Strength,
		Weight:      req.Strength, // Weight 默认等于 Strength
	}

	log.Printf("[Handler] UpdateRelation Calling Service with relation.ID=%s", relation.ID)

	// 构建命名空间
	namespace := types.NameSpace{
		TenantID: strconv.FormatInt(tenantID, 10),
		KBID:     kbID,
	}

	log.Printf("[Handler] UpdateRelation Namespace: TenantID=%s, KBID=%s", namespace.TenantID, namespace.KBID)

	// 更新关系，直接返回更新后的数据
	updatedRelation, err := h.graphService.UpdateRelation(c.Request.Context(), namespace, relation)
	if err != nil {
		log.Printf("[Handler] UpdateRelation ERROR from Service: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Handler] UpdateRelation Service returned: updatedRelation=%v", updatedRelation != nil)

	// 如果找不到关系（返回 nil），返回 404
	if updatedRelation == nil {
		log.Printf("[Handler] UpdateRelation WARNING: Service returned nil, relation not found")
		c.JSON(404, gin.H{"error": "relation not found"})
		return
	}

	log.Printf("[Handler] UpdateRelation SUCCESS: returning ID=%s, Source=%q, Target=%q, Type=%q, Strength=%f, Weight=%f",
		updatedRelation.ID, updatedRelation.Source, updatedRelation.Target, updatedRelation.Type, updatedRelation.Strength, updatedRelation.Weight)

	c.JSON(200, gin.H{
		"message": "success",
		"data":    updatedRelation,
	})
}

// DeleteNode 删除单个节点
// DELETE /api/v1/knowledge-bases/:id/graph/nodes/:nodeId
func (h *GraphHandler) DeleteNode(c *gin.Context) {
	kbID := c.Param("id")
	nodeID := c.Param("nodeId")

	log.Printf("[Handler] DeleteNode START: kb_id=%s, node_id=%s", kbID, nodeID)

	// 使用 middleware 辅助函数获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		log.Printf("[Handler] DeleteNode ERROR: unauthorized, missing tenant_id")
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 构建命名空间
	namespace := types.NameSpace{
		TenantID: strconv.FormatInt(tenantID, 10),
		KBID:     kbID,
	}

	// 删除节点
	if err := h.graphService.DeleteNode(c.Request.Context(), namespace, nodeID); err != nil {
		log.Printf("[Handler] DeleteNode ERROR: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Handler] DeleteNode SUCCESS: node_id=%s", nodeID)

	c.JSON(200, gin.H{
		"message": "success",
	})
}

// DeleteRelation 删除单个关系
// DELETE /api/v1/knowledge-bases/:id/graph/relations/:relationId
func (h *GraphHandler) DeleteRelation(c *gin.Context) {
	kbID := c.Param("id")
	relationID := c.Param("relationId")

	log.Printf("[Handler] DeleteRelation START: kb_id=%s, relation_id=%s", kbID, relationID)

	// 使用 middleware 辅助函数获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		log.Printf("[Handler] DeleteRelation ERROR: unauthorized, missing tenant_id")
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 构建命名空间
	namespace := types.NameSpace{
		TenantID: strconv.FormatInt(tenantID, 10),
		KBID:     kbID,
	}

	// 删除关系
	if err := h.graphService.DeleteRelation(c.Request.Context(), namespace, relationID); err != nil {
		log.Printf("[Handler] DeleteRelation ERROR: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Handler] DeleteRelation SUCCESS: relation_id=%s", relationID)

	c.JSON(200, gin.H{
		"message": "success",
	})
}

// GetRelationTypes 获取关系类型选项
// GET /api/v1/knowledge-bases/:id/graph/relation-types
func (h *GraphHandler) GetRelationTypes(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "success",
		"data":    types.RelationTypeOptions(),
	})
}

// DeleteGraph 删除知识库图谱
// DELETE /api/v1/knowledge-bases/:id/graph
func (h *GraphHandler) DeleteGraph(c *gin.Context) {
	kbID := c.Param("id")

	// 使用 middleware 辅助函数获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 构建命名空间
	namespace := types.NameSpace{
		TenantID: strconv.FormatInt(tenantID, 10),
		KBID:     kbID,
	}

	// 删除图谱
	if err := h.graphService.DeleteGraph(c.Request.Context(), []types.NameSpace{namespace}); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
	})
}
