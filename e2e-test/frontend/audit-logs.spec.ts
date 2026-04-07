import { test, expect, Page } from './test-helpers'

test.describe('Audit Logs Page', () => {
  test('should display audit logs page title', async ({ page }) => {
    await page.goto('/audit-logs')
    await expect(page.locator('h4:has-text("审计日志")')).toBeVisible()
  })

  test('should display search form', async ({ page }) => {
    await page.goto('/audit-logs')
    await expect(page.locator('label:has-text("用户 ID")')).toBeVisible()
    await expect(page.locator('label:has-text("实例")')).toBeVisible()
  })
})
