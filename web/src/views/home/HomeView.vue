<template>
  <div class="home-view">
    <!-- 头部欢迎区域 -->
    <header class="home-header">
      <div class="header-content">
        <h1 class="title">
          <span class="gradient-text">Link</span>
          <span class="subtitle">智能知识管理平台</span>
        </h1>
        <p class="description">
          集成大模型对话、知识库管理、知识图谱、Agent 智能体的一站式 AI 平台
        </p>
      </div>
    </header>

    <!-- 功能卡片区域 -->
    <div class="features-section">
      <div class="section-title">
        <h2>核心功能</h2>
        <p>探索平台的强大功能，提升您的工作效率</p>
      </div>

      <div class="features-grid">
        <!-- 聊天功能 -->
        <div class="feature-card" @click="navigateTo('/chat')">
          <div class="card-icon">
            <el-icon :size="48"><ChatLineRound /></el-icon>
          </div>
          <h3>智能对话</h3>
          <p>与大模型进行自然流畅的对话，支持多种模型切换，支持 Agent 模式自动调用工具完成复杂任务</p>
          <ul class="feature-list">
            <li>多模型支持</li>
            <li>上下文记忆</li>
            <li>Agent 智能体</li>
            <li>深度研究模式</li>
          </ul>
          <div class="card-action">
            <span>开始对话</span>
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>

        <!-- 知识库功能 -->
        <div class="feature-card" @click="navigateTo('/knowledge')">
          <div class="card-icon">
            <el-icon :size="48"><Document /></el-icon>
          </div>
          <h3>知识库管理</h3>
          <p>上传和管理您的文档，智能分块索引，支持 RAG 检索增强生成，让 AI 更懂您的知识</p>
          <ul class="feature-list">
            <li>多格式文档支持</li>
            <li>智能分块</li>
            <li>向量检索</li>
            <li>重排序优化</li>
          </ul>
          <div class="card-action">
            <span>管理知识库</span>
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>

        <!-- 知识图谱功能 -->
        <div class="feature-card" @click="navigateTo('/graphs')">
          <div class="card-icon">
            <el-icon :size="48"><Share /></el-icon>
          </div>
          <h3>知识图谱</h3>
          <p>可视化展示知识实体关系，支持 Graph RAG，通过图谱增强检索，提供更精准的答案</p>
          <ul class="feature-list">
            <li>实体关系抽取</li>
            <li>图谱可视化</li>
            <li>图数据库存储</li>
            <li>图谱检索增强</li>
          </ul>
          <div class="card-action">
            <span>查看图谱</span>
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>

        <!-- Agent 管理 -->
        <div class="feature-card" @click="navigateTo('/agent')">
          <div class="card-icon">
            <el-icon :size="48"><User /></el-icon>
          </div>
          <h3>Agent 管理</h3>
          <p>创建和管理智能 Agent，配置工具链，让 AI 自动完成复杂任务，提升工作效率</p>
          <ul class="feature-list">
            <li>Agent 配置</li>
            <li>工具调用</li>
            <li>任务编排</li>
            <li>多 Agent 协作</li>
          </ul>
          <div class="card-action">
            <span>管理 Agent</span>
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>

        <!-- 大模型测评 -->
        <div class="feature-card" @click="navigateTo('/evaluation')">
          <div class="card-icon">
            <el-icon :size="48"><DataAnalysis /></el-icon>
          </div>
          <h3>大模型测评</h3>
          <p>全面评估大模型性能，支持多维度测评指标，帮助您选择最适合的模型</p>
          <ul class="feature-list">
            <li>自动评测</li>
            <li>多维度指标</li>
            <li>对比分析</li>
            <li>报告导出</li>
          </ul>
          <div class="card-action">
            <span>开始测评</span>
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>

        <!-- 系统设置 -->
        <div class="feature-card" @click="navigateTo('/settings')">
          <div class="card-icon">
            <el-icon :size="48"><Setting /></el-icon>
          </div>
          <h3>系统设置</h3>
          <p>配置模型参数、API 信息、系统偏好等，个性化定制您的平台体验</p>
          <ul class="feature-list">
            <li>模型配置</li>
            <li>API 管理</li>
            <li>偏好设置</li>
            <li>租户管理</li>
          </ul>
          <div class="card-action">
            <span>前往设置</span>
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>
      </div>
    </div>

    <!-- 快速开始区域 -->
    <div class="quick-start-section">
      <div class="section-title">
        <h2>快速开始</h2>
        <p>三步即可开始使用 Link 平台</p>
      </div>

      <div class="steps">
        <div class="step-item">
          <div class="step-number">1</div>
          <div class="step-content">
            <h4>配置模型</h4>
            <p>前往设置页面，配置您的大模型 API 密钥</p>
          </div>
        </div>
        <div class="step-arrow">
          <el-icon><ArrowRight /></el-icon>
        </div>
        <div class="step-item">
          <div class="step-number">2</div>
          <div class="step-content">
            <h4>创建知识库</h4>
            <p>上传文档，系统将自动分块并建立索引</p>
          </div>
        </div>
        <div class="step-arrow">
          <el-icon><ArrowRight /></el-icon>
        </div>
        <div class="step-item">
          <div class="step-number">3</div>
          <div class="step-content">
            <h4>开始对话</h4>
            <p>与 AI 对话，体验 RAG 和 Agent 带来的智能交互</p>
          </div>
        </div>
      </div>

      <div class="quick-actions">
        <el-button type="primary" size="large" @click="navigateTo('/settings')">
          <el-icon><Setting /></el-icon>
          配置模型
        </el-button>
        <el-button size="large" @click="navigateTo('/chat')">
          <el-icon><ChatLineRound /></el-icon>
          开始对话
        </el-button>
      </div>
    </div>

    <!-- 技术亮点 -->
    <div class="tech-section">
      <div class="section-title">
        <h2>技术亮点</h2>
      </div>
      <div class="tech-tags">
        <el-tag v-for="tech in technologies" :key="tech" size="large">{{ tech }}</el-tag>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import {
  ChatLineRound,
  Document,
  Share,
  User,
  DataAnalysis,
  Setting,
  ArrowRight
} from '@element-plus/icons-vue'

const router = useRouter()

const technologies = [
  '大模型 LLM',
  'RAG 检索增强',
  '知识图谱',
  'Agent 智能体',
  '向量数据库',
  '图数据库',
  '深度搜索',
  '多模型支持'
]

function navigateTo(path: string) {
  router.push(path)
}
</script>

<style scoped>
.home-view {
  max-width: 1400px;
  margin: 0 auto;
  padding: 40px 20px;
}

/* 头部区域 */
.home-header {
  text-align: center;
  padding: 60px 20px;
  margin-bottom: 40px;
}

.header-content {
  max-width: 800px;
  margin: 0 auto;
}

.title {
  font-size: 56px;
  font-weight: 700;
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.gradient-text {
  background: linear-gradient(135deg, var(--color-primary-light), var(--color-secondary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-size: 72px;
}

.subtitle {
  font-size: 32px;
  color: var(--color-text-primary);
  font-weight: 500;
}

.description {
  font-size: 18px;
  color: var(--color-text-secondary);
  line-height: 1.6;
  max-width: 600px;
  margin: 0 auto;
}

/* 区域标题 */
.section-title {
  text-align: center;
  margin-bottom: 40px;
}

.section-title h2 {
  font-size: 32px;
  color: var(--color-text-primary);
  margin-bottom: 12px;
}

.section-title p {
  font-size: 16px;
  color: var(--color-text-muted);
}

/* 功能卡片区域 */
.features-section {
  margin-bottom: 60px;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(380px, 1fr));
  gap: 24px;
}

.feature-card {
  position: relative;
  background: rgba(0, 0, 0, 0.25);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-lg);
  padding: 32px;
  cursor: pointer;
  transition: all var(--transition-base);
}

.feature-card:hover {
  transform: translateY(-4px);
  border-color: var(--color-primary);
  box-shadow: var(--shadow-glow);
}

.card-icon {
  width: 80px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(255, 126, 179, 0.2), rgba(122, 252, 255, 0.2));
  border-radius: var(--radius-md);
  margin-bottom: 20px;
  color: var(--color-primary-light);
}

.feature-card h3 {
  font-size: 24px;
  color: var(--color-text-primary);
  margin-bottom: 12px;
}

.feature-card > p {
  font-size: 15px;
  color: var(--color-text-secondary);
  line-height: 1.6;
  margin-bottom: 20px;
  min-height: 48px;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0 0 24px 0;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.feature-list li {
  font-size: 13px;
  color: var(--color-text-muted);
  padding: 4px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 20px;
}

.card-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--color-primary-light);
  font-weight: 500;
  padding-top: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

/* 快速开始区域 */
.quick-start-section {
  background: rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-lg);
  padding: 40px;
  margin-bottom: 60px;
}

.steps {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 40px;
}

.step-item {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
  min-width: 200px;
}

.step-number {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
  border-radius: 50%;
  font-size: 20px;
  font-weight: 600;
  color: white;
  flex-shrink: 0;
}

.step-content h4 {
  font-size: 16px;
  color: var(--color-text-primary);
  margin-bottom: 4px;
}

.step-content p {
  font-size: 14px;
  color: var(--color-text-muted);
}

.step-arrow {
  color: var(--color-text-muted);
  font-size: 20px;
}

.quick-actions {
  display: flex;
  justify-content: center;
  gap: 16px;
  flex-wrap: wrap;
}

/* 技术亮点 */
.tech-section {
  text-align: center;
  padding: 40px 0;
}

.tech-tags {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 12px;
}

.tech-tags .el-tag {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: var(--color-text-secondary);
  padding: 8px 20px;
  font-size: 15px;
}

/* 响应式 */
@media (max-width: 1024px) {
  .features-grid {
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  }
}

@media (max-width: 768px) {
  .home-view {
    padding: 20px 16px;
  }

  .title {
    font-size: 40px;
  }

  .gradient-text {
    font-size: 48px;
  }

  .subtitle {
    font-size: 24px;
  }

  .features-grid {
    grid-template-columns: 1fr;
  }

  .steps {
    flex-direction: column;
  }

  .step-arrow {
    transform: rotate(90deg);
  }

  .quick-actions {
    flex-direction: column;
  }

  .quick-actions .el-button {
    width: 100%;
  }
}
</style>
