import { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Form,
  Input,
  DatePicker,
  Button,
  Space,
  Typography,
  Tag,
  message,
} from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { apiClient, AuditLog } from '../api'

const { Title } = Typography
const { RangePicker } = DatePicker

export default function AuditLogs() {
  const [loading, setLoading] = useState(false)
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [form] = Form.useForm()
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20 })

  useEffect(() => {
    fetchLogs()
  }, [pagination])

  const fetchLogs = async (params?: {
    userId?: string
    instance?: string
    startTime?: string
    endTime?: string
  }) => {
    setLoading(true)
    try {
      const response = await apiClient.getAuditLogs({
        ...params,
        page: pagination.current,
        pageSize: pagination.pageSize,
      })
      if (response.success && response.data) {
        setLogs(response.data.logs)
        setTotal(response.data.total)
      } else {
        message.error(response.error?.message || '获取审计日志失败')
      }
    } catch {
      message.error('请求失败')
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = () => {
    const values = form.getFieldsValue()
    const params: Record<string, string> = {}
    
    if (values.userId) params.userId = values.userId
    if (values.instance) params.instance = values.instance
    if (values.timeRange && values.timeRange[0] && values.timeRange[1]) {
      params.startTime = values.timeRange[0].toISOString()
      params.endTime = values.timeRange[1].toISOString()
    }

    setPagination({ ...pagination, current: 1 })
    fetchLogs(params)
  }

  const handleReset = () => {
    form.resetFields()
    setPagination({ ...pagination, current: 1 })
    fetchLogs()
  }

  const columns = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (ts: string) => dayjs(ts).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '用户',
      dataIndex: 'userId',
      key: 'userId',
      width: 120,
    },
    {
      title: '实例',
      dataIndex: 'instance',
      key: 'instance',
      width: 120,
    },
    {
      title: 'SQL',
      dataIndex: 'sql',
      key: 'sql',
      ellipsis: true,
      render: (sql: string) => (
        <Typography.Text
          style={{ fontFamily: 'monospace', fontSize: 12 }}
          ellipsis={{ tooltip: sql }}
        >
          {sql}
        </Typography.Text>
      ),
    },
    {
      title: '耗时',
      dataIndex: 'executionTime',
      key: 'executionTime',
      width: 100,
      render: (time: number) => `${(time / 1000000).toFixed(2)}ms`,
    },
    {
      title: '影响行数',
      dataIndex: 'rowsAffected',
      key: 'rowsAffected',
      width: 100,
    },
    {
      title: '状态',
      dataIndex: 'success',
      key: 'success',
      width: 80,
      render: (success: boolean) => (
        <Tag color={success ? 'success' : 'error'}>
          {success ? '成功' : '失败'}
        </Tag>
      ),
    },
    {
      title: '错误信息',
      dataIndex: 'errorMessage',
      key: 'errorMessage',
      ellipsis: true,
      render: (msg: string) =>
        msg ? (
          <Typography.Text type="danger" ellipsis={{ tooltip: msg }}>
            {msg}
          </Typography.Text>
        ) : (
          '-'
        ),
    },
  ]

  return (
    <div>
      <Title level={4} style={{ marginBottom: 16 }}>
        审计日志
      </Title>

      <Card size="small" style={{ marginBottom: 16 }}>
        <Form form={form} layout="inline">
          <Form.Item name="userId" label="用户 ID">
            <Input placeholder="输入用户 ID" style={{ width: 150 }} />
          </Form.Item>
          <Form.Item name="instance" label="实例">
            <Input placeholder="输入实例名" style={{ width: 150 }} />
          </Form.Item>
          <Form.Item name="timeRange" label="时间范围">
            <RangePicker showTime style={{ width: 350 }} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button
                type="primary"
                icon={<SearchOutlined />}
                onClick={handleSearch}
              >
                搜索
              </Button>
              <Button icon={<ReloadOutlined />} onClick={handleReset}>
                重置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      <Card>
        <Table
          dataSource={logs}
          columns={columns}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 条记录`,
            onChange: (page, pageSize) =>
              setPagination({ current: page, pageSize }),
          }}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  )
}
