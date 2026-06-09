# Frontend Glass Archive Redesign Design

## Goal

Redesign `web/frontend` into a refined personal technical blog with a "glass archive" identity: private knowledge archive, light glass interface layers, strong long-form reading, and full light/dark/system theme support.

## Product Reading

Nostalgia is a personal technical blog, not a marketing site or admin dashboard. The frontend should feel like a carefully maintained private archive: articles are indexed records, categories are archive labels, activity and metadata are supporting cabinet drawers, and the article page is a quiet reading desk.

The UI must keep the existing user flows:

- Browse latest articles.
- Browse by category.
- Search articles.
- Read article details by ID or slug.
- View article metadata, likes, views, read time, summary, and outdated warning.
- Like an article.
- Read and write comments.
- Log in and register.

## Scope

First implementation phase covers only `web/frontend`.

Included:

- Frontend UI stack cleanup.
- Global design tokens and theme runtime.
- Navigation and theme switcher.
- Home article list.
- Sidebar modules: GitHub activity and category list.
- Search and category result views.
- Article detail reading page.
- Article metadata, summary, copyright block, like action, comments baseline styling.
- Loading, empty, error, and responsive states for the touched surfaces.

Excluded:

- Go backend changes.
- `web/backend` admin redesign.
- API contract changes.
- Dark-mode redesign of backend management UI.
- New content authoring features.

## Technology Direction

Use route 1:

- `Tailwind CSS` for styling and token application.
- `shadcn-vue` for owned Vue component implementations.
- `Reka UI` for accessible primitives.
- Vue 3, Vite, Pinia, Vue Router, Axios remain.
- CKEditor, Prism, CalHeatmap remain unless a specific implementation issue requires a targeted replacement.

PrimeVue must be removed from the frontend UI surface:

- Remove PrimeVue component usage from `web/frontend`.
- Remove PrimeVue theme setup from `web/frontend/src/main.ts`.
- Replace PrimeVue components currently used in frontend views and components:
  - Toast
  - Paginator
  - Skeleton
  - Breadcrumb
  - Button
  - Panel
  - InputText
  - InputGroup
  - InputGroupAddon
  - FloatLabel
  - Menu
  - DataView
  - Badge
  - Divider
  - Message
  - ConfirmDialog
  - Tooltip directive
- Keep `primeicons` only if the implementation chooses to retain the icon family temporarily. Prefer replacing icons with a single non-Prime icon family if practical during phase one.

## Visual Identity

The design language is "private glass archive."

Use:

- Archive indexing, labels, dates, record rows, metadata strips, fine dividers.
- Light glass for navigation, side panels, metadata, search, theme switcher, and floating actions.
- Solid reading surfaces for article body, code blocks, comments, and forms.
- Refined spacing and typography over decorative effects.

Avoid:

- Full-page glassmorphism.
- Purple/blue neon gradients.
- Deep blue-gray AI-style dark glass.
- Warm beige parchment nostalgia.
- Heavy card grids and repeated icon-card patterns.
- Glass behind long-form body text.

## Color And Material System

### Shared Principles

- One primary accent: copper green / archive green.
- Use neutral ink, graphite, slate, and cold paper tones.
- Use borders and material layering before shadows.
- Glass is an interface material, not a reading material.
- Body text must meet WCAG AA contrast in both themes.

### Light Theme

Light theme is the default reading-first environment.

- Page background: cold white / very pale neutral gray.
- Body text: near-ink.
- Secondary text: readable slate gray.
- Accent: deep archive green.
- Glass layer: translucent white with subtle blur and fine white/gray border.
- Reading layer: solid white or near-white with a quiet border.
- Code layer: solid light code background with clear token colors.

### Dark Theme

Dark theme should feel like a night archive room, not a generic dark-tech UI.

- Page background: neutral graphite / smoked ink, not blue-purple.
- Body text: soft off-white.
- Secondary text: warm-neutral or neutral gray with enough contrast.
- Accent: archive green with slightly higher luminance than light mode.
- Glass layer: smoked graphite glass, mostly neutral black/gray, with subtle transparency.
- Border highlight: low-luminance neutral line plus very restrained green focus ring.
- Reading layer: solid dark graphite, slightly lighter than page background.
- Code layer: solid dark code surface with no transparency.

Dark glass rule: do not use a saturated deep-blue-gray panel as the default material. If hue is needed, bias it toward neutral graphite with a small green-tinted edge only on focus or active states.

## Theme Runtime

Theme mode has three states:

- `system`
- `light`
- `dark`

Behavior:

- Store selected mode in `localStorage` under `nostalgia-theme-mode`.
- If no stored mode exists, default to `system`.
- If mode is `system`, follow `window.matchMedia('(prefers-color-scheme: dark)')`.
- If the system preference changes while mode is `system`, update the applied theme.
- Apply the resolved theme to the document using `data-theme="light"` or `data-theme="dark"`.
- Components consume CSS variables and Tailwind tokens, not hard-coded color literals.

Theme switcher:

- Desktop: compact three-state segmented control in the navigation area.
- Mobile: compact icon button or menu item that exposes the same three choices.
- The selected state must be visible and keyboard accessible.

## Page Design

### App Shell

The app shell uses a full-height layout:

- Sticky or near-sticky glass navigation at top.
- Main content with constrained max width.
- Footer simplified into a quiet archive footer.

Navigation:

- Logo remains the primary home link.
- Search remains in the nav, but should become a refined archive search field.
- User menu uses accessible popover/dropdown behavior from Reka UI or owned shadcn-vue components.
- Mobile navigation should collapse into clear menus without clipped dropdowns.

### Home And Results Layout

The home page keeps the two-column blog shape on desktop:

- Main column: article index.
- Side column: activity and categories.

Desktop:

- Main column should read like an archive index, not a stack of generic cards.
- Article items include title, summary, cover, category, author, date, likes, views, and read time if available.
- Metadata should be grouped into compact archive labels.
- Cover images should be stable in aspect ratio and should not cause layout shift.
- Side modules use light glass panels with strong internal hierarchy.

Mobile:

- Single-column layout.
- Sidebar modules move below article list or become collapsible sections.
- Search and theme controls remain reachable without crowding the nav.

### Article Detail

Article detail is the most important reading surface.

Layout:

- Reading width target: 680-760px.
- Page can include a subtle metadata header above the article.
- The body uses a solid reading surface.
- Summary has an archive note treatment without a thick side stripe.
- Outdated warning is visible but not loud.
- Reading progress remains but uses a refined accent line.

Content:

- CKEditor content must be styled with readable headings, paragraphs, lists, tables, blockquotes, images, and links.
- Code blocks use a solid background and clear syntax contrast.
- Inline code is distinct but not noisy.
- Blockquotes use full border/background treatment rather than a thick colored left stripe.

Actions:

- Like action can use a small glass action well or solid compact control.
- Copyright block should feel like an archive license record, with calm hierarchy.
- Comment editor and comment list should use solid readable surfaces.

### Search And Category Views

Replace breadcrumbs with a custom archive trail:

- Home icon/link.
- Current mode: search or category.
- Query/category name.

Results reuse the home article index component.

### Login And Register

Phase one does not need a complete visual overhaul of login/register unless needed to remove PrimeVue. The replacement should be usable, theme-aware, and consistent enough:

- Solid form panel.
- Clear labels above fields.
- Accessible error and disabled states.
- Login/register links remain.

## Component Architecture

Create owned frontend UI components rather than styling everything inline.

Suggested units:

- Theme runtime composable: handles mode, resolved theme, storage, and system listener.
- Theme switcher component: renders the three-state control.
- App navigation component: owns nav layout, search, user menu, mobile menus.
- UI primitives: button, input, badge, panel, skeleton, dialog, toast, dropdown, pagination.
- Article index item component: one article row/card.
- Article index component: data rendering, loading/empty/error states, pagination slot.
- Archive trail component: replaces PrimeVue breadcrumb.
- Sidebar panel component: glass panel wrapper for GitHub activity and categories.
- Article reading layout component or CSS module for CKEditor content.

Each unit should have clear responsibilities and avoid large files doing layout, fetching, and styling all at once when reasonable.

## Data Flow

Keep existing API services:

- `listArticle`
- `searchArticles`
- `getArticle`
- `getArticleBySlug`
- `incrementArticleLikes`
- `incrementArticleViews`
- `listCategories`
- `listComments`
- existing user and comment store actions

UI refactor should not change API payload shapes.

The article list still receives `categoryId` and `keyword` props.

The theme runtime is independent from auth, articles, comments, and routing.

## States And Error Handling

Every touched surface should have explicit states:

- Loading: skeletons matching the final layout.
- Empty: calm archive empty state, not a generic inbox icon.
- Error: readable inline or toast feedback.
- Disabled: visible but not low-contrast.
- Focus: keyboard-visible and theme-aware.
- Reduced motion: transitions remain minimal or disabled under `prefers-reduced-motion`.

Toast replacement must support at least:

- success
- info
- warning
- error

Confirm dialog replacement must support deleting comments.

## Accessibility

- Theme switcher must be keyboard accessible.
- Dropdowns and dialogs should use Reka UI primitives or equivalent accessible behavior.
- Form labels must be real labels, not placeholder-only labels.
- Text contrast must pass in light and dark themes.
- Interactive controls need visible focus states.
- Mobile menus must be operable without hover.
- Motion must respect `prefers-reduced-motion`.

## Testing And Verification

At minimum:

- Run `npm run type-check` in `web/frontend`.
- Run `npm run build` in `web/frontend`.
- Run frontend dev server and inspect:
  - home page
  - article detail
  - search results
  - category results
  - login
  - register
  - mobile viewport
  - light, dark, and system modes
- Verify localStorage theme persistence by changing theme, refreshing, and checking resolved theme.
- Verify no PrimeVue imports remain in `web/frontend/src`.
- Verify package dependencies no longer include PrimeVue frontend UI dependencies after migration.

## Design Approval Notes

Confirmed by user:

- Design direction: private archive.
- Keep reading experience friendly and refined.
- Use light glass material.
- Support light, dark, and system theme modes.
- Persist theme mode in browser localStorage.
- Fully remove PrimeVue from frontend.
- Use shadcn-vue + Reka UI + Tailwind CSS as the new frontend UI foundation.

