import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	webServer: {
		command: 'npm run preview',
		port: 4173,
		reuseExistingServer: !process.env.CI
	},
	testDir: 'tests',
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
		}
	],
	use: {
		baseURL: 'http://localhost:4173'
	}
});
