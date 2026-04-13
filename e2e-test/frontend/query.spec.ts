import { test, expect, Page } from './test-helpers'

test.describe('Query Page', () => {
  test('should display query page elements', async ({ page }) => {
    await page.goto('/query')
    await expect(page.locator('text=SQL 编辑器')).toBeVisible()
    await expect(page.locator('text=查询结果')).toBeVisible()
  })

  test('should show warning when executing without selecting instance', async ({ page }) => {
    await page.goto('/query')
    await page.click('button:has-text("执行")')
    
    await expect(page.locator('.ant-message-warning')).toBeVisible()
  })
})
