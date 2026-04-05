import { useState } from 'react'
import { Form, Input, Button, Card, message, Typography } from 'antd'
import { useAuthStore } from '../stores/authStore'
import { apiClient } from '../api'

const { Title, Text } = Typography

interface LoginForm {
  token: string
}

export default function Login() {
  const [loading, setLoading] = useState(false)
  const setAuth = useAuthStore((state) => state.setAuth)
  const [form] = Form.useForm()

  const handleSubmit = async (values: LoginForm) => {
    setLoading(true)
    try {
      const result = await apiClient.verifyAndLogin(values.token)
      if (result.success && result.userId && result.role && result.expiresAt) {
        const expiresAt = new Date(result.expiresAt).getTime()
        setAuth(values.token, result.userId, result.role, expiresAt)
        message.success('登录成功')
        window.location.href = '/dashboard'
      } else {
        message.error(result.error || 'Token 无效')
      }
    } catch {
      message.error('验证请求失败，请检查网络连接')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        height: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Card style={{ width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.15)' }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <Title level={2} style={{ marginBottom: 8 }}>
            MystiSql
          </Title>
          <Text type="secondary">
            数据库访问网关
          </Text>
        </div>
        <Form
          form={form}
          onFinish={handleSubmit}
          layout="vertical"
        >
          <Form.Item
            name="token"
            label="Token"
            rules={[{ required: true, message: '请输入 Token' }]}
          >
            <Input.TextArea
              placeholder="请输入 Token"
              autoSize={{ minRows: 2, maxRows: 4 }}
              size="large"
            />
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              size="large"
            >
              验证并登录
            </Button>
          </Form.Item>
        </Form>
        <Text
          type="secondary"
          style={{ display: 'block', textAlign: 'center' }}
        >
          使用 <code>mystisql auth bootstrap</code> 命令获取管理员 Token
        </Text>
      </Card>
    </div>
  )
}
