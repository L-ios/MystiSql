import { test, expect, Page } from './test-helpers'

test.describe('Dashboard Page', () => {
  test('should display dashboard title', async ({ page }) => {
    await expect(page.locator('h4:has-text("仪表盘")')).toBeVisible()
  })

  test('should display statistics cards', async ({ page }) => {
    await expect(page.locator('text=总实例数')).toBeVisible()
    await expect(page.locator('text=健康实例')).toBeVisible()
    await expect(page.locator('text=异常实例')).toBeVisible()
    await expect(page.locator('text=在线时长')).toBeVisible()
  })

  test('should display instance status overview', async ({ page }) => {
    await expect(page.locator('text=实例状态概览')).toBeVisible()
  })
})
