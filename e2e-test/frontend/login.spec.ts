import { test, expect } from '@playwright/test'

test.describe('Login Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
  })

  test('should display login form', async ({ page }) => {
    await expect(page.locator('h2:has-text("MystiSql")')).toBeVisible()
    await expect(page.locator('text=数据库访问网关')).toBeVisible()
    await expect(page.locator('[data-testid="login-button"]')).toBeVisible()
    await expect(page.locator('textarea[placeholder="请输入 Token"]')).toBeVisible()
  })

  test('should show validation error for empty token', async ({ page }) => {
    await page.click('[data-testid="login-button"]')

    await expect(page.locator('text=请输入 Token')).toBeVisible()
  })

  test('should login successfully with valid token', async ({ page }) => {
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

    await page.fill('textarea[placeholder="请输入 Token"]', 'valid-test-token')

    await Promise.all([
      page.waitForURL('**/dashboard', { timeout: 30000 }),
      page.click('[data-testid="login-button"]'),
    ])

    await expect(page).toHaveURL(/.*dashboard/)
  })

  test('should show error message for invalid token', async ({ page }) => {
    await page.route('**/api/v1/auth/token/info', route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: {
            code: 'INVALID_TOKEN',
            message: '无效的 Token',
          },
        }),
      })
    })

    await page.fill('textarea[placeholder="请输入 Token"]', 'invalid-token')
    await page.click('[data-testid="login-button"]')

    await expect(page.locator('text=无效的 Token')).toBeVisible({ timeout: 10000 })
  })

  test('should redirect to login when accessing protected route without auth', async ({ page }) => {
    await page.goto('/dashboard')
    await page.waitForURL('**/login**', { timeout: 10000 })
    await expect(page).toHaveURL(/.*login/)
  })

  test('should have proper form accessibility', async ({ page }) => {
    const button = page.locator('[data-testid="login-button"]')
    await expect(button).toBeVisible()
    await expect(button).toBeEnabled()
    await expect(button).toHaveText(/登录/)
  })

  test('should handle network errors gracefully', async ({ page }) => {
    await page.route('**/api/v1/auth/token/info', route => route.abort())

    await page.fill('textarea[placeholder="请输入 Token"]', 'some-token')
    await page.click('[data-testid="login-button"]')

    // 网络错误时应停留在登录页并给出提示，而不是崩溃
    await page.waitForTimeout(2000)
    await expect(page).toHaveURL(/.*login/)
  })
})
