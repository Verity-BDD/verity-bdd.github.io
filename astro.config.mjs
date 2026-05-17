// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

// https://astro.build/config
export default defineConfig({
  site: "https://verity-bdd.github.io",
  base: "/",
  redirects: {
    "/": "/en/",
  },
  integrations: [
    starlight({
      customCss: ["./src/styles/custom.css"],
      title: "Verity BDD",
      description: "Screenplay Pattern testing framework for Go",
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/verity-bdd/verity-bdd",
        },
      ],
      defaultLocale: "en",
      locales: {
        en: { label: "English", lang: "en" },
        // ru: { label: 'Русский', lang: 'ru' },
      },
      sidebar: [
        {
          label: "Get Started",
          translations: { ru: "С чего начать" },
          items: [{ autogenerate: { directory: "get_started" } }],
        },
        {
          label: "Core Concepts",
          translations: { ru: "Ключевые особенности" },
          items: [{ autogenerate: { directory: "core_concepts" } }],
        },
        {
          label: "Recipies",
          translations: { ru: "Рецепты" },
          items: [{ autogenerate: { directory: "guides" } }],
        },
        {
          label: "Examples",
          translations: { ru: "Примеры" },
          items: [{ autogenerate: { directory: "examples" } }],
        },
        {
          label: "API Reference",
          translations: { ru: "API Справочник" },
          items: [{ autogenerate: { directory: "api" } }],
        },
      ],
    }),
  ],
});
