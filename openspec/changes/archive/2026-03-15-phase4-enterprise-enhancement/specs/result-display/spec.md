## ADDED Requirements

### Requirement: 结果集表格展示
系统 SHALL 以表格形式展示查询结果。

#### Scenario: 显示结果表格
- **WHEN** 查询返回数据
- **THEN** 以表格形式显示列和行

#### Scenario: 自动调整列宽
- **WHEN** 显示结果表格
- **THEN** 列宽根据内容自动调整

### Requirement: 结果集分页
系统 SHALL 支持大结果集分页展示。

#### Scenario: 分页显示
- **WHEN** 结果超过 100 行
- **THEN** 自动分页显示

#### Scenario: 翻页操作
- **WHEN** 用户点击下一页
- **THEN** 显示下一页数据

### Requirement: 结果集排序
系统 SHALL 支持结果集排序。

#### Scenario: 点击列头排序
- **WHEN** 用户点击列头
- **THEN** 结果按该列排序

#### Scenario: 切换升序降序
- **WHEN** 用户再次点击同一列头
- **THEN** 切换排序方向

### Requirement: 结果集导出
系统 SHALL 支持导出查询结果。

#### Scenario: 导出为 CSV
- **WHEN** 用户点击导出 CSV
- **THEN** 下载 CSV 格式文件

#### Scenario: 导出为 JSON
- **WHEN** 用户点击导出 JSON
- **THEN** 下载 JSON 格式文件

### Requirement: 大结果集虚拟滚动
系统 SHALL 使用虚拟滚动处理大结果集。

#### Scenario: 虚拟滚动
- **WHEN** 结果集超过 10000 行
- **THEN** 使用虚拟滚动，仅渲染可见行
