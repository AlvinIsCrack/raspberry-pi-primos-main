import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				runes: ({ filename }) => filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			adapter: adapter({
				pages: '../src/web/static',
				assets: '../src/web/static',
				fallback: 'index.html',
				precompress: false,
				strict: true
			})
		})
	],
	server: {
		proxy: {
			'/api': {
				target: 'http://127.0.0.1:8080',
				changeOrigin: true,
				// Required for streaming Server-Sent Events without buffering
				ws: false,
				configure: (proxy) => {
					proxy.on('proxyRes', (proxyRes) => {
						if (proxyRes.headers['content-type'] === 'text/event-stream') {
							proxyRes.headers['cache-control'] = 'no-cache';
							proxyRes.headers['connection'] = 'keep-alive';
						}
					});
				}
			}
		}
	}
});
