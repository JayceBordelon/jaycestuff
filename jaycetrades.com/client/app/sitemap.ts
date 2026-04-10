import type { MetadataRoute } from "next";

export default function sitemap(): MetadataRoute.Sitemap {
	return [
		{
			url: "https://jaycetrades.com",
			changeFrequency: "daily",
			priority: 1.0,
		},
		{
			url: "https://jaycetrades.com/history",
			changeFrequency: "daily",
			priority: 0.8,
		},
		{
			url: "https://jaycetrades.com/terms",
			changeFrequency: "monthly",
			priority: 0.3,
		},
		{
			url: "https://jaycetrades.com/faq",
			changeFrequency: "monthly",
			priority: 0.5,
		},
	];
}
