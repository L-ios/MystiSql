import { test, expect, Page } from '@playwright/test'

test.describe('Login Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
  })

  test('should display login form', async ({ page }) => {
    await expect(page.locator('text=MystiSql')).toBeVisible()
    await expect(page.locator('text=数据库访问网关')).toBeVisible()
    await expect(page.locator('[data-testid="user-id-input"]')).toBeVisible()
    await expect(page.locator('[data-testid="role-input"]')).toBeVisible()
    await expect(page.locator('[data-testid="login-button"]')).toBeVisible()
  })

  test('should show validation error for empty user ID', async ({ page }) => {
    await page.fill('[data-testid="user-id-input"]', '')
    await page.fill('[data-testid="role-input"]', '')
    await page.click('[data-testid="login-button"]')
    
    await expect(page.locator('text=请输入用户 ID')).toBeVisible()
  })

  test('should login successfully with valid credentials', async ({ page }) => {
    await page.route('**/api/v1/auth/token', route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            token: 'test-token-12345',
            userId: 'test-user',
            role: 'admin',
            expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
          }
        })
      })
    })

    await page.fill('[data-testid="user-id-input"]', 'test-user')
    await page.fill('[data-testid="role-input"]', 'admin')
    
    await Promise.all([
      page.waitForURL('**/dashboard', { timeout: 30000 }),
      page.click('[data-testid="login-button"]')
    ])
    
    await expect(page).toHaveURL(/.*dashboard/)
  })

  test('should show error message for invalid credentials', async ({ page }) => {
    await page.route('**/api/v1/auth/token', route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: {
            code: 'UNAUTHORIZED',
            message: 'Invalid credentials'
          }
        })
      })
    })

    await page.fill('[data-testid="user-id-input"]', 'invalid-user')
    await page.fill('[data-testid="role-input"]', 'invalid-role')
    
    await page.click('[data-testid="login-button"]')
    
    await expect(page.locator('.ant-message')).toBeVisible({ timeout: 10000 })
  })

  test('should redirect to login when accessing protected route without auth', async ({ page }) => {
    await page.goto('/dashboard')
    
    await expect(page).toHaveURL(/.*login/)
  })

  test('should have proper form accessibility', async ({ page }) => {
    const userIdInput = page.locator('[data-testid="user-id-input"]')
    const roleInput = page.locator('[data-testid="role-input"]')
    
    await expect(userIdInput).toHaveAttribute('type', 'text')
    await expect(roleInput).toHaveAttribute('type', 'text')
    
    const loginButton = page.locator('[data-testid="login-button"]')
    await expect(loginButton).toHaveAttribute('type', 'submit')
  })

  test('should handle network errors gracefully', async ({ page }) => {
    await page.route('**/api/v1/auth/token', route => {
      route.abort('failed')
    })

    await page.fill('[data-testid="user-id-input"]', 'test-user')
    await page.fill('[data-testid="role-input"]', 'admin')
    
    await page.click('[data-testid="login-button"]')
    
    await expect(page.locator('.ant-message')).toBeVisible({ timeout: 10000 })
  })
})
