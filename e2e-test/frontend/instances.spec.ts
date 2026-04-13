import { test, expect, Page } from './test-helpers'

test.describe('Instances Page', () => {
  test('should display instances page title', async ({ page }) => {
    await page.goto('/instances')
    await expect(page.locator('h4:has-text("实例管理")')).toBeVisible()
  })

  test('should display instances content', async ({ page }) => {
    await page.goto('/instances')
    await expect(page.locator('.ant-card')).toBeVisible()
  })
})
