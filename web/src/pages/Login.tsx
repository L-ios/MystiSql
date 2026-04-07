import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Form, Input, Button, Card, message, Typography } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useAuthStore } from '../stores/authStore'
import { apiClient } from '../api'

const { Title } = Typography

interface LoginForm {
  userId: string
  role: string
}

export default function Login() {
  const [loading, setLoading] = useState(false)
  const setAuth = useAuthStore((state) => state.setAuth)
  const [form] = Form.useForm()
  const navigate = useNavigate()

  const handleSubmit = async (values: LoginForm) => {
    setLoading(true)
    try {
      const response = await apiClient.login(values.userId, values.role)
      if (response.success && response.data) {
        const expiresAt = new Date(response.data.expiresAt).getTime()
        setAuth(
          response.data.token,
          response.data.userId,
          response.data.role,
          expiresAt
        )
        message.success('登录成功')
        navigate('/dashboard')
      } else {
        message.error(response.error?.message || '登录失败')
      }
    } catch {
      message.error('登录请求失败，请检查网络连接')
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
          <Typography.Text type="secondary">
            数据库访问网关
          </Typography.Text>
        </div>
        <Form
          form={form}
          onFinish={handleSubmit}
          layout="vertical"
          initialValues={{ role: 'readonly' }}
        >
          <Form.Item
            name="userId"
            label="用户 ID"
            rules={[{ required: true, message: '请输入用户 ID' }]}
          >
            <Input
              data-testid="user-id-input"
              prefix={<UserOutlined />}
              placeholder="请输入用户 ID"
              size="large"
            />
          </Form.Item>
          <Form.Item
            name="role"
            label="角色"
            rules={[{ required: true, message: '请输入角色' }]}
          >
            <Input
              data-testid="role-input"
              prefix={<LockOutlined />}
              placeholder="请输入角色（如：admin、readonly）"
              size="large"
            />
          </Form.Item>
          <Form.Item>
            <Button
              data-testid="login-button"
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              size="large"
            >
              登录
            </Button>
          </Form.Item>
        </Form>
        <Typography.Text
          type="secondary"
          style={{ display: 'block', textAlign: 'center' }}
        >
          使用 Token 认证登录
        </Typography.Text>
      </Card>
    </div>
  )
}
