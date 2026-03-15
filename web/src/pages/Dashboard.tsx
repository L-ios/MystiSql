import { useEffect } from 'react'
import { Card, Row, Col, Statistic, Typography, Spin, Tag } from 'antd'
import {
  DatabaseOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import { useInstanceStore } from '../stores/instanceStore'
import { apiClient } from '../api'

const { Title } = Typography

export default function Dashboard() {
  const { instances, setInstances, loading, setLoading } = useInstanceStore()

  useEffect(() => {
    fetchInstances()
  }, [])

  const fetchInstances = async () => {
    setLoading(true)
    try {
      const response = await apiClient.getInstances()
      setInstances(response.instances)
    } catch {
      console.error('Failed to fetch instances')
    } finally {
      setLoading(false)
    }
  }

  const healthyCount = instances.filter((i) => i.status === 'healthy').length
  const unhealthyCount = instances.filter((i) => i.status !== 'healthy').length

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 100 }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <Title level={4} style={{ marginBottom: 24 }}>
        仪表盘
      </Title>
      <Row gutter={16}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总实例数"
              value={instances.length}
              prefix={<DatabaseOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="健康实例"
              value={healthyCount}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="异常实例"
              value={unhealthyCount}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: unhealthyCount > 0 ? '#cf1322' : '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="在线时长"
              value={0}
              suffix="分钟"
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Title level={5} style={{ marginTop: 32, marginBottom: 16 }}>
        实例状态概览
      </Title>
      <Row gutter={[16, 16]}>
        {instances.map((instance) => (
          <Col span={8} key={instance.name}>
            <Card size="small">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <Typography.Text strong>{instance.name}</Typography.Text>
                  <br />
                  <Typography.Text type="secondary">
                    {instance.type} · {instance.host}:{instance.port}
                  </Typography.Text>
                </div>
                <Tag
                  color={instance.status === 'healthy' ? 'success' : 'error'}
                  icon={
                    instance.status === 'healthy' ? (
                      <CheckCircleOutlined />
                    ) : (
                      <CloseCircleOutlined />
                    )
                  }
                >
                  {instance.status}
                </Tag>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  )
}
