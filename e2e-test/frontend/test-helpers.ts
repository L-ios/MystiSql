import { test as base, expect, Page as PlaywrightPage } from '@playwright/test'

export type { Page } from '@playwright/test'

// 通用的 API mock：让前端页面在无后端环境下可测
async function mockApis(page: PlaywrightPage) {
  // 登录验证
  await page.route('**/api/v1/auth/token/info', route => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        userId: 'test-user',
        role: 'admin',
        expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
      }),
    })
  })

  // 实例列表
  await page.route('**/api/v1/instances', route => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        instances: [
          { name: 'test-mysql', type: 'mysql', host: '127.0.0.1', port: 3306, status: 'healthy' },
        ],
      }),
    })
  })

  // 服务健康（含 uptime）
  await page.route('**/health', route => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        status: 'healthy',
        version: 'test',
        uptime: 3600,
        timestamp: new Date().toISOString(),
      }),
    })
  })
}

export const test = base.extend({
  page: async ({ page }, use) => {
    await mockApis(page)

    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    await page.fill('textarea[placeholder="请输入 Token"]', 'test-token-12345')

    await Promise.all([
      page.waitForURL('**/dashboard', { timeout: 30000 }),
      page.click('[data-testid="login-button"]'),
    ])

    await page.waitForLoadState('domcontentloaded')

    await use(page)
  },
})

export { expect }
