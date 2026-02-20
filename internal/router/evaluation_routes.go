package router

import (
	"github.com/gin-gonic/gin"
)

// SetupEvaluationRoutes 设置测评路由
func SetupEvaluationRoutes(
	api *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
	createEvalHandler,
	getEvalHandler,
	listEvalHandler,
	getEvalByIDHandler,
	deleteEvalHandler,
	createDatasetHandler,
	listDatasetsHandler gin.HandlerFunc,
) {
	// 测评任务路由
	eval := api.Group("/evaluation")
	eval.Use(authMiddleware)
	{
		eval.POST("", createEvalHandler) // 创建测评任务
		eval.GET("", getEvalHandler)     // 获取测评结果（通过task_id查询）
	}

	// 测评任务管理路由
	evals := api.Group("/evaluations")
	evals.Use(authMiddleware)
	{
		evals.GET("", listEvalHandler)          // 列出测评任务
		evals.GET("/:id", getEvalByIDHandler)   // 获取单个测评任务
		evals.DELETE("/:id", deleteEvalHandler) // 删除测评任务
	}

	// 数据集管理路由
	datasets := api.Group("/datasets")
	datasets.Use(authMiddleware)
	{
		datasets.POST("", createDatasetHandler) // 创建数据集
		datasets.GET("", listDatasetsHandler)   // 列出数据集
	}
}
