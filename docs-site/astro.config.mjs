// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://docs.runaic.com',
	integrations: [
		starlight({
			title: 'aic CLI',
			description: 'Official command-line interface for AIC.',
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/learners-superpumped/aic' },
			],
			sidebar: [
				{
					label: 'Guides',
					items: [
						{ label: 'Installation', slug: 'guides/installation' },
						{ label: 'Quick start', slug: 'guides/quick-start' },
					],
				},
				{
					label: 'Command reference',
					items: [{ autogenerate: { directory: 'reference' } }],
				},
			],
		}),
	],
});
