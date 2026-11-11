---
paths:
  - "web/app/**"
---

# Web host

- Hand-written TypeScript, no UI framework, no router, no state library. The host owns the DOM, input, rendering of blocks, scrolling, the mobile keyboard and themes; the shell owns everything else.
- Build DOM with `createElement` and `textContent`. Never `innerHTML`, `insertAdjacentHTML` or `document.write` with shell output, file content, command arguments, the URL hash or feed data.
- Links from shell output are validated (`https:`, `mailto:` or in-site paths) before they become `href`.
- Honour `prefers-reduced-motion` and `prefers-color-scheme`; every action works from the keyboard; output lives in an `aria-live` region.
- Prettier: 120 columns, single quotes, no trailing commas. ESLint flat config. `@`-aliases are declared in both `tsconfig.json` and `webpack.config.js`.
- No code comments.
