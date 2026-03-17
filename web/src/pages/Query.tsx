import { useState, useCallback } from 'react'
import {
  Card,
  Select,
  Button,
  Spin,
  Alert,
  Table,
  Typography,
  Empty,
  message,
  Space,
  Dropdown,
  Tabs,
} from 'antd'
import {
  PlayCircleOutlined,
  DownloadOutlined,
  ClearOutlined,
  HistoryOutlined,
  CopyOutlined,
} from '@ant-design/icons'
import Editor from '@monaco-editor/react'
import { useInstanceStore } from '../stores/instanceStore'
import { useQueryStore } from '../stores/queryStore'
import { apiClient, QueryResultData } from '../api'

const { Text, Paragraph } = Typography

export default function Query() {
  const { instances, currentInstance, setCurrentInstance } = useInstanceStore()
  const { currentSql, setCurrentSql, history, addHistory, clearHistory } = useQueryStore()
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<QueryResultData | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [execTime, setExecTime] = useState<number>(0)

  const handleExecute = useCallback(async () => {
    if (!currentInstance) {
      message.warning('请先选择实例')
      return
    }
    if (!currentSql.trim()) {
      message.warning('请输入 SQL 语句')
      return
    }

    setLoading(true)
    setError(null)
    setResult(null)

    try {
      const response = await apiClient.query({
        instance: currentInstance,
        sql: currentSql,
      })

      setExecTime(response.executionTime)

      if (response.success && response.data) {
        setResult(response.data)
        addHistory({
          sql: currentSql,
          instance: currentInstance,
          success: true,
        })
        message.success(`查询成功，返回 ${response.data.rowCount} 行`)
      } else {
        setError(response.error?.message || '查询失败')
        addHistory({
          sql: currentSql,
          instance: currentInstance,
          success: false,
        })
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '请求失败'
      setError(errorMsg)
      addHistory({
        sql: currentSql,
        instance: currentInstance,
        success: false,
      })
    } finally {
      setLoading(false)
    }
  }, [currentInstance, currentSql, addHistory])

  const handleExport = useCallback((format: 'csv' | 'json') => {
    if (!result) return

    let content: string
    let filename: string
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')

    if (format === 'json') {
      content = JSON.stringify(result.rows, null, 2)
      filename = `query-result-${timestamp}.json`
    } else {
      const headers = result.columns.map((c) => c.name).join(',')
      const rows = result.rows
        .map((row) =>
          result.columns
            .map((c) => {
              const val = row[c.name]
              if (val === null || val === undefined) return ''
              if (typeof val === 'string' && val.includes(',')) {
                return `"${val.replace(/"/g, '""')}"`
              }
              return String(val)
            })
            .join(',')
        )
        .join('\n')
      content = `${headers}\n${rows}`
      filename = `query-result-${timestamp}.csv`
    }

    const blob = new Blob([content], {
      type: format === 'json' ? 'application/json' : 'text/csv',
    })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
    message.success('导出成功')
  }, [result])

  const handleCopy = useCallback(() => {
    if (!result) return
    navigator.clipboard.writeText(JSON.stringify(result.rows, null, 2))
    message.success('已复制到剪贴板')
  }, [result])

  const handleHistoryClick = useCallback((sql: string) => {
    setCurrentSql(sql)
  }, [setCurrentSql])

  const columns = result?.columns.map((col) => ({
    title: col.name,
    dataIndex: col.name,
    key: col.name,
    ellipsis: true,
    sorter: (a: Record<string, unknown>, b: Record<string, unknown>) => {
      const valA = a[col.name]
      const valB = b[col.name]
      if (valA === null || valA === undefined) return -1
      if (valB === null || valB === undefined) return 1
      if (typeof valA === 'number' && typeof valB === 'number') {
        return valA - valB
      }
      return String(valA).localeCompare(String(valB))
    },
    render: (value: unknown) => {
      if (value === null) return <Text type="secondary">NULL</Text>
      if (value === undefined) return ''
      return String(value)
    },
  })) || []

  const historyItems = history.slice(0, 20).map((item, index) => ({
    key: index,
    label: (
      <div style={{ maxWidth: 300 }}>
        <Text ellipsis style={{ display: 'block' }}>
          {item.sql}
        </Text>
        <Text type="secondary" style={{ fontSize: 12 }}>
          {new Date(item.timestamp).toLocaleString()} · {item.instance}
        </Text>
      </div>
    ),
    onClick: () => handleHistoryClick(item.sql),
  }))

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space>
          <Select
            style={{ width: 200 }}
            placeholder="选择实例"
            value={currentInstance}
            onChange={setCurrentInstance}
            options={instances.map((i) => ({
              label: i.name,
              value: i.name,
            }))}
          />
          <Button
            type="primary"
            icon={<PlayCircleOutlined />}
            onClick={handleExecute}
            loading={loading}
          >
            执行 (Ctrl+Enter)
          </Button>
          <Dropdown menu={{ items: historyItems }} placement="bottomLeft">
            <Button icon={<HistoryOutlined />}>
              历史记录 ({history.length})
            </Button>
          </Dropdown>
          <Button icon={<ClearOutlined />} onClick={() => clearHistory()}>
            清空历史
          </Button>
        </Space>
      </Card>

      <div style={{ flex: 1, display: 'flex', gap: 16, minHeight: 0 }}>
        <Card
          size="small"
          title="SQL 编辑器"
          style={{ flex: 1, display: 'flex', flexDirection: 'column' }}
          styles={{ body: { flex: 1, padding: 0 } }}
        >
          <Editor
            height="100%"
            defaultLanguage="sql"
            value={currentSql}
            onChange={(value) => setCurrentSql(value || '')}
            theme="vs"
            options={{
              minimap: { enabled: false },
              fontSize: 14,
              lineNumbers: 'on',
              wordWrap: 'on',
              automaticLayout: true,
              scrollBeyondLastLine: false,
            }}
            onMount={(editor, monaco) => {
              editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
                handleExecute()
              })
            }}
          />
        </Card>

        <Card
          size="small"
          title={
            <span>
              查询结果
              {execTime > 0 && (
                <Text type="secondary" style={{ marginLeft: 8, fontSize: 12 }}>
                  耗时: {(execTime / 1000000).toFixed(2)}ms
                </Text>
              )}
            </span>
          }
          extra={
            result && (
              <Space>
                <Button
                  size="small"
                  icon={<CopyOutlined />}
                  onClick={handleCopy}
                >
                  复制
                </Button>
                <Dropdown
                  menu={{
                    items: [
                      { key: 'csv', label: '导出 CSV', onClick: () => handleExport('csv') },
                      { key: 'json', label: '导出 JSON', onClick: () => handleExport('json') },
                    ],
                  }}
                >
                  <Button size="small" icon={<DownloadOutlined />}>
                    导出
                  </Button>
                </Dropdown>
              </Space>
            )
          }
          style={{ flex: 1, display: 'flex', flexDirection: 'column' }}
          styles={{ body: { flex: 1, overflow: 'auto', padding: 12 } }}
        >
          {loading && (
            <div style={{ textAlign: 'center', padding: 50 }}>
              <Spin size="large" />
            </div>
          )}
          {error && <Alert type="error" message={error} />}
          {!loading && !error && !result && (
            <Empty description="执行查询以查看结果" />
          )}
          {!loading && !error && result && (
            <Tabs
              items={[
                {
                  key: 'table',
                  label: `表格 (${result.rowCount} 行)`,
                  children: (
                    <Table
                      dataSource={result.rows}
                      columns={columns}
                      rowKey={(_, index) => `row-${index}`}
                      size="small"
                      scroll={{ x: 'max-content', y: 400 }}
                      pagination={{
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (total) => `共 ${total} 行`,
                        defaultPageSize: 50,
                        pageSizeOptions: ['20', '50', '100', '200'],
                      }}
                    />
                  ),
                },
                {
                  key: 'json',
                  label: 'JSON',
                  children: (
                    <Paragraph>
                      <pre style={{ margin: 0, maxHeight: 400, overflow: 'auto' }}>
                        {JSON.stringify(result.rows, null, 2)}
                      </pre>
                    </Paragraph>
                  ),
                },
              ]}
            />
          )}
        </Card>
      </div>
    </div>
  )
}
