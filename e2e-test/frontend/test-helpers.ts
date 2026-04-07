import { test as base, expect, Page as PlaywrightPage } from '@playwright/test'

export type { Page } from '@playwright/test'

export const test = base.extend({
  page: async ({ page }, use) => {
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

    await page.goto('/login')
    await page.waitForLoadState('networkidle')
    
    await page.fill('[data-testid="user-id-input"]', 'test-user')
    await page.fill('[data-testid="role-input"]', 'admin')
    
    await Promise.all([
      page.waitForURL('**/dashboard', { timeout: 30000 }),
      page.click('[data-testid="login-button"]')
    ])
    
    await page.waitForLoadState('domcontentloaded')
    
    await use(page)
  },
})

export { expect }
