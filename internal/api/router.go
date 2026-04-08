package api

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"ai-writer/internal/config"
	"ai-writer/internal/api/handler"
	"ai-writer/internal/api/middleware"
	"ai-writer/internal/llm"
	"ai-writer/internal/store"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// 初始化存储
	jsonStore := store.NewJSONStore(".")
	handler.InitStore(jsonStore)
	handler.InitConfig(cfg)

	// 初始化向量存储
	vectorDBClient, err := store.NewSQLiteVectorDB(cfg.GetVectorDBPath())
	if err != nil {
		log.Printf("Warning: Failed to initialize vector DB: %v", err)
	} else {
		handler.InitVectorDB(vectorDBClient)
	}

	// 初始化 embedding 客户端
	embeddingClient := llm.NewEmbeddingClient(
		cfg.Embedding.Provider,
		cfg.Embedding.BaseURL,
		cfg.Embedding.APIKey,
		cfg.Embedding.Model,
	)

	// 如果是 Python provider，等待服务就绪
	if cfg.Embedding.Provider == "python" || cfg.Embedding.Provider == "" {
		pythonClient, ok := embeddingClient.(*llm.PythonEmbeddingClient)
		if ok {
			log.Println("Waiting for Python embedding service...")
			if err := pythonClient.WaitForReady(30 * time.Second); err != nil {
				log.Fatalf("Embedding service not ready: %v", err)
			}
			log.Println("Python embedding service ready")
		}
	}

	handler.InitEmbeddingClient(embeddingClient)

	// 中间件
	router.Use(middleware.Cors())
	router.Use(middleware.Recovery())

	// 静态文件（前端）
	router.Static("/assets", "./web/dist/assets")
	router.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})

	// API 路由组
	api := router.Group("/api")
	{
		// 书籍管理
		books := api.Group("/books")
		{
			books.GET("", handler.ListBooks)
			books.POST("", handler.CreateBook)
			books.GET("/:id", handler.GetBook)
			books.PUT("/:id", handler.UpdateBook)
			books.DELETE("/:id", handler.DeleteBook)

			// 章节管理 - 使用 :id 作为书籍参数
			bookChapters := books.Group("/:id/chapters")
			{
				bookChapters.GET("", handler.ListChapters)
				bookChapters.POST("", handler.CreateChapter)
				bookChapters.GET("/:chapter_id", handler.GetChapter)
				bookChapters.PUT("/:chapter_id", handler.UpdateChapter)
				bookChapters.DELETE("/:chapter_id", handler.DeleteChapter)
				bookChapters.GET("/:chapter_id/content", handler.GetChapterContent)
				bookChapters.PUT("/:chapter_id/content", handler.UpdateChapterContent)
			}

			// 设定管理
			bookSettings := books.Group("/:id/settings")
			{
				bookSettings.GET("/worldview", handler.GetWorldView)
				bookSettings.PUT("/worldview", handler.UpdateWorldView)
				bookSettings.GET("/characters", handler.ListCharacters)
				bookSettings.POST("/characters", handler.CreateCharacter)
				bookSettings.PUT("/characters/:char_id", handler.UpdateCharacter)
				bookSettings.DELETE("/characters/:char_id", handler.DeleteCharacter)
				bookSettings.GET("/items", handler.ListItems)
				bookSettings.POST("/items", handler.CreateItem)
				bookSettings.PUT("/items/:item_id", handler.UpdateItem)
				bookSettings.DELETE("/items/:item_id", handler.DeleteItem)
				bookSettings.GET("/locations", handler.ListLocations)
				bookSettings.POST("/locations", handler.CreateLocation)
				bookSettings.PUT("/locations/:loc_id", handler.UpdateLocation)
				bookSettings.DELETE("/locations/:loc_id", handler.DeleteLocation)
			}

			// 因果链
			bookCausal := books.Group("/:id/causal-chain")
			{
				bookCausal.GET("", handler.GetCausalChains)
				bookCausal.POST("", handler.CreateCausalEvent)
				bookCausal.PUT("/:event_id", handler.UpdateCausalEvent)
			}

			// 伏笔
			bookForeshadow := books.Group("/:id/foreshadows")
			{
				bookForeshadow.GET("", handler.ListForeshadows)
				bookForeshadow.POST("", handler.CreateForeshadow)
				bookForeshadow.PUT("/:fid", handler.UpdateForeshadow)
				bookForeshadow.POST("/:fid/resolve", handler.ResolveForeshadow)
				bookForeshadow.GET("/warnings", handler.GetForeshadowWarnings)
			}

			// 时间线
			bookTimeline := books.Group("/:id/timeline")
			{
				bookTimeline.GET("", handler.GetTimeline)
				bookTimeline.GET("/threads", handler.GetNarrativeThreads)
				bookTimeline.POST("/threads", handler.CreateNarrativeThread)
			}

			// 图谱
			bookGraph := books.Group("/:id/graph")
			{
				bookGraph.GET("", handler.GetKnowledgeGraph)
				bookGraph.GET("/echarts", handler.GetEChartsData)
			}

			// 世界状态审计
			bookSync := books.Group("/:id/sync")
			{
				bookSync.POST("/extract-all", handler.SyncExtractAll)
				bookSync.GET("/pending-graphs", handler.SyncGetPendingGraphs)
				bookSync.POST("/apply-graphs", handler.SyncApplyGraphs)
			}

			// 分析
			bookAnalysis := books.Group("/:id/analysis")
			{
				bookAnalysis.POST("/run", handler.AnalysisRun)
				bookAnalysis.GET("/reports", handler.AnalysisGetReports)
			}
		}

		// AI 写作（需要 LLM 配置）
		ai := api.Group("/ai")
		{
			ai.POST("/generate", handler.AIGenerate)
			ai.POST("/generate/stream", handler.AIGenerateStream)
			ai.GET("/review", handler.AIReview)
			ai.POST("/review", handler.AIReview)
			ai.POST("/audit", handler.AIAudit)
			ai.POST("/rewrite", handler.AIRewrite)
			ai.POST("/continue", handler.AIContinue)
		}

		// 批量生成
		batch := api.Group("/batch")
		{
			batch.POST("/generate", handler.BatchGenerate)
			batch.GET("/status", handler.BatchStatus)
			batch.DELETE("/reset", handler.BatchReset)
		}

		// 导出
		api.GET("/books/:id/export/:format", handler.ExportBook)

		// 状态同步
		sync := api.Group("/sync")
		{
			sync.POST("/extract", handler.SyncExtract)
			sync.GET("/pending", handler.SyncPending)
			sync.POST("/apply", handler.SyncApply)
			sync.POST("/reject", handler.SyncReject)
		}

		// 智能工具箱
		toolbox := api.Group("/toolbox")
		{
			toolbox.POST("/naming", handler.ToolNaming)
			toolbox.POST("/character", handler.ToolCharacter)
			toolbox.POST("/conflict", handler.ToolConflict)
			toolbox.POST("/scene", handler.ToolScene)
			toolbox.POST("/goldfinger", handler.ToolGoldfinger)
			toolbox.POST("/title", handler.ToolTitle)
			toolbox.POST("/synopsis", handler.ToolSynopsis)
			toolbox.POST("/twist", handler.ToolTwist)
			toolbox.POST("/dialogue", handler.ToolDialogue)
		}

		// 架构师
		architect := api.Group("/architect")
		{
			architect.POST("/generate", handler.ArchitectGenerate)
			architect.POST("/fission", handler.ArchitectFission)
			architect.GET("/strategies", handler.ArchitectStrategies)
		}

		// 拆书分析
		analysisGroup := api.Group("/analysis")
		{
			analysisGroup.POST("/parse", handler.AnalysisParse)
			analysisGroup.POST("/analyze", handler.AnalysisAnalyze)
		}

		// 系统设置
		system := api.Group("/system")
		{
			system.GET("/config", handler.GetConfig)
			system.PUT("/config", handler.UpdateConfig)
			system.GET("/prompts", handler.GetPrompts)
			system.PUT("/prompts", handler.UpdatePrompts)
			system.GET("/billing", handler.GetBillingStats)
			system.GET("/goals", handler.GetWritingGoals)
			system.PUT("/goals", handler.UpdateWritingGoals)
			system.GET("/ollama/models", handler.GetOllamaModels)
		}

		// 向量存储
		vector := api.Group("/vector")
		{
			vector.POST("/index", handler.VectorIndexBook)
			vector.POST("/index/chapter", handler.VectorIndexChapter)
			vector.POST("/search", handler.VectorSearch)
			vector.GET("/status", handler.VectorStatus)
			vector.DELETE("/index/:book_name", handler.VectorDeleteBook)
		}
	}

	return router
}