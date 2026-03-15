import { useEffect, useState } from 'react'
import {
  Card,
  Table,
  Tag,
  Button,
  Space,
  Typography,
  Modal,
  Descriptions,
  Spin,
  message,
  Badge,
} from 'antd'
import {
  ReloadOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons'
import { useInstanceStore } from '../stores/instanceStore'
import { apiClient, PoolStats } from '../api'

const { Title } = Typography

interface InstanceDetail {
  name: string
  health: string
  poolStats?: PoolStats
}

export default function Instances() {
  const { instances, setInstances, loading, setLoading } = useInstanceStore()
  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedInstance, setSelectedInstance] = useState<InstanceDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)

  useEffect(() => {
    fetchInstances()
  }, [])

  const fetchInstances = async () => {
    setLoading(true)
    try {
      const response = await apiClient.getInstances()
      setInstances(response.instances)
    } catch {
      message.error('获取实例列表失败')
    } finally {
      setLoading(false)
    }
  }

  const showInstanceDetail = async (name: string) => {
    setDetailLoading(true)
    setDetailVisible(true)
    setSelectedInstance({ name, health: 'unknown' })

    try {
      const [healthRes, poolRes] = await Promise.all([
        apiClient.getInstanceHealth(name),
        apiClient.getPoolStats(name),
      ])
      setSelectedInstance({
        name,
        health: healthRes.status,
        poolStats: poolRes.stats,
      })
    } catch {
      message.error('获取实例详情失败')
    } finally {
      setDetailLoading(false)
    }
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string) => <Typography.Text strong>{name}</Typography.Text>,
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag>{type.toUpperCase()}</Tag>,
    },
    {
      title: '地址',
      key: 'address',
      render: (_: unknown, record: { host: string; port: number }) =>
        `${record.host}:${record.port}`,
    },
    {
      title: '数据库',
      dataIndex: 'database',
      key: 'database',
      render: (db: string) => db || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Badge
          status={status === 'healthy' ? 'success' : 'error'}
          text={status === 'healthy' ? '健康' : '异常'}
        />
      ),
    },
    {
      title: '标签',
      dataIndex: 'labels',
      key: 'labels',
      render: (labels: Record<string, string>) =>
        labels
          ? Object.entries(labels).map(([k, v]) => (
              <Tag key={k} color="blue">
                {k}={v}
              </Tag>
            ))
          : '-',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: { name: string }) => (
        <Button
          type="link"
          icon={<EyeOutlined />}
          onClick={() => showInstanceDetail(record.name)}
        >
          详情
        </Button>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Title level={4} style={{ margin: 0 }}>
          实例管理
        </Title>
        <Button icon={<ReloadOutlined />} onClick={fetchInstances} loading={loading}>
          刷新
        </Button>
      </div>

      <Card>
        <Table
          dataSource={instances}
          columns={columns}
          rowKey="name"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 个实例`,
          }}
        />
      </Card>

      <Modal
        title="实例详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={600}
      >
        {detailLoading ? (
          <div style={{ textAlign: 'center', padding: 50 }}>
            <Spin />
          </div>
        ) : selectedInstance ? (
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="实例名称" span={2}>
                {selectedInstance.name}
              </Descriptions.Item>
              <Descriptions.Item label="健康状态">
                {selectedInstance.health === 'healthy' ? (
                  <Tag icon={<CheckCircleOutlined />} color="success">
                    健康
                  </Tag>
                ) : (
                  <Tag icon={<CloseCircleOutlined />} color="error">
                    异常
                  </Tag>
                )}
              </Descriptions.Item>
            </Descriptions>

            {selectedInstance.poolStats && (
              <>
                <Title level={5}>连接池状态</Title>
                <Descriptions bordered column={2} size="small">
                  <Descriptions.Item label="最大连接数">
                    {selectedInstance.poolStats.maxOpenConnections}
                  </Descriptions.Item>
                  <Descriptions.Item label="当前连接数">
                    {selectedInstance.poolStats.openConnections}
                  </Descriptions.Item>
                  <Descriptions.Item label="使用中">
                    {selectedInstance.poolStats.inUse}
                  </Descriptions.Item>
                  <Descriptions.Item label="空闲">
                    {selectedInstance.poolStats.idle}
                  </Descriptions.Item>
                  <Descriptions.Item label="等待次数">
                    {selectedInstance.poolStats.waitCount}
                  </Descriptions.Item>
                  <Descriptions.Item label="等待时间">
                    {(selectedInstance.poolStats.waitDuration / 1000000).toFixed(2)}ms
                  </Descriptions.Item>
                </Descriptions>
              </>
            )}
          </Space>
        ) : null}
      </Modal>
    </div>
  )
}
