<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1>{{ isLogin ? t('auth.login') : t('auth.register') }}</h1>
        <p class="subtitle">Link - AI Knowledge Platform</p>
      </div>

      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        class="login-form"
        @submit.prevent="handleSubmit"
      >
        <!-- 用户名 (注册时显示) -->
        <el-form-item v-if="!isLogin" prop="username">
          <el-input
            v-model="formData.username"
            :placeholder="t('auth.usernamePlaceholder')"
            size="large"
            clearable
          >
            <template #prefix>
              <el-icon><User /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 邮箱 -->
        <el-form-item prop="email">
          <el-input
            v-model="formData.email"
            :placeholder="t('auth.emailPlaceholder')"
            size="large"
            clearable
          >
            <template #prefix>
              <el-icon><Message /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 密码 -->
        <el-form-item prop="password">
          <el-input
            v-model="formData.password"
            type="password"
            :placeholder="t('auth.passwordPlaceholder')"
            size="large"
            show-password
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 确认密码 (注册时显示) -->
        <el-form-item v-if="!isLogin" prop="confirmPassword">
          <el-input
            v-model="formData.confirmPassword"
            type="password"
            :placeholder="t('auth.confirmPasswordPlaceholder')"
            size="large"
            show-password
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <!-- 记住我 (登录时显示) -->
        <el-form-item v-if="isLogin">
          <el-checkbox v-model="formData.rememberMe">{{ t('auth.rememberMe') }}</el-checkbox>
        </el-form-item>

        <!-- 提交按钮 -->
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            native-type="submit"
            class="submit-btn"
          >
            {{ loading ? t('common.loading') : (isLogin ? t('auth.login') : t('auth.register')) }}
          </el-button>
        </el-form-item>

        <!-- 切换登录/注册 -->
        <div class="switch-mode">
          <span v-if="isLogin">
            {{ t('auth.noAccount') }}
            <el-link type="primary" @click="toggleMode">{{ t('auth.goToRegister') }}</el-link>
          </span>
          <span v-else>
            {{ t('auth.hasAccount') }}
            <el-link type="primary" @click="toggleMode">{{ t('auth.goToLogin') }}</el-link>
          </span>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Message, Lock } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// 是否登录模式
const isLogin = ref(true)
const loading = ref(false)
const formRef = ref<FormInstance>()

// 表单数据
const formData = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  rememberMe: false
})

// 验证规则
const rules: FormRules = {
  username: [
    { required: true, message: t('auth.usernameRequired'), trigger: 'blur' },
    { min: 2, max: 20, message: '用户名长度在2到20个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: t('auth.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('auth.invalidEmail'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: t('auth.passwordRequired'), trigger: 'blur' },
    { min: 6, message: t('auth.passwordTooShort'), trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: t('auth.passwordRequired'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value !== formData.password) {
          callback(new Error(t('auth.passwordNotMatch')))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 切换登录/注册模式
function toggleMode() {
  isLogin.value = !isLogin.value
  formRef.value?.clearValidate()
  // 清空确认密码
  formData.confirmPassword = ''
}

// 处理提交
async function handleSubmit() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      if (isLogin.value) {
        // 登录
        const success = await authStore.login({
          email: formData.email,
          password: formData.password
        })

        if (success) {
          ElMessage.success(t('auth.loginSuccess'))

          // 登录成功，跳转到原来的页面或平台首页
          const redirect = (route.query.redirect as string) || '/'
          await router.push(redirect)
        } else {
          ElMessage.error('登录失败，请检查用户名和密码')
        }
      } else {
        // 注册
        const success = await authStore.register({
          username: formData.username,
          email: formData.email,
          password: formData.password
        })

        if (success) {
          ElMessage.success(t('auth.registerSuccess'))

          // 注册成功，切换到登录模式
          isLogin.value = true
          formData.confirmPassword = ''
          formRef.value?.clearValidate()
        } else {
          ElMessage.error('注册失败，请稍后重试')
        }
      }
    } catch (error: any) {
      // 错误已经在 request.ts 中处理并显示，这里只需要打印日志
      console.error('Auth error:', error)
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-box {
  width: 400px;
  padding: 40px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  font-size: 28px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
}

.subtitle {
  color: #999;
  font-size: 14px;
}

.login-form {
  margin-top: 24px;
}

.submit-btn {
  width: 100%;
  margin-top: 8px;
}

.switch-mode {
  text-align: center;
  margin-top: 16px;
  font-size: 14px;
  color: #666;
}

.switch-mode .el-link {
  font-size: 14px;
}
</style>
