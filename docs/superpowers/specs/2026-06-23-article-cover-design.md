# Article Cover Standardization Design

## Goal

Standardize article cover images so they feel intentional across public article lists, public article detail pages, backend previews, and social sharing metadata. The first version should improve visual consistency without introducing a cropper, backend image processing pipeline, or new storage fields.

## Decisions

- Use one canonical cover image per article, stored in the existing `article.cover` field.
- Use `16:9` as the authoring and display standard.
- Recommend `1600x900` for normal use and `1920x1080` for high-resolution covers.
- Warn when an image is smaller than `1200x675`.
- Warn when the image ratio is far from `16:9`, but do not block upload, save, or publish.
- Show the cover at the top of the public article detail page, above the title.
- Continue using the existing cover URL as `og:image` and `twitter:image`.
- Do not add a cropper, focal point metadata, generated derivatives, or backend image processing in this version.

## Current Context

The backend and frontend already expose one cover URL on article records. The frontend SEO metadata builder already maps `article.cover` to Open Graph and Twitter image metadata. Backend article preview and cover preview styling already lean toward `16:9`, while the public article list currently uses a more contained image treatment that can make covers feel boxed-in.

The public article detail page currently renders article content through the shared `ArticleReader` component, but the detail route does not yet show the article cover. This design makes the cover a first-class reading-page element while keeping the data model unchanged.

## Scope

In scope:

- Add a consistent cover display rule for public article detail pages.
- Align public article list cover rendering with the same cover visual language.
- Keep backend article preview and cover preview aligned with public rendering.
- Add backend cover guidance and warnings for dimensions and aspect ratio.
- Add multi-surface backend preview for detail, list, and social contexts.
- Keep SEO cover metadata behavior intact.

Out of scope:

- New database columns for cover variants.
- Backend image resizing, compression, WebP/AVIF conversion, or derivative generation.
- Manual cropper UI.
- Focal point controls.
- Historical cover migration.
- Hard rejection of non-`16:9` images.

## Cover Standard

The product standard is:

```text
Aspect ratio: 16:9
Recommended size: 1600x900
High-resolution size: 1920x1080
Minimum recommended size: 1200x675
```

The UI should present these as recommendations, not as hard upload requirements. This keeps the system friendly to existing content, temporary images, and edge cases where the owner intentionally chooses a different composition.

## Public Article Detail

When an article has a cover, render it above the article title and metadata. The cover should use the same content width as the article reader surface and a fixed `16:9` container.

Rendering rules:

- Use `object-fit: cover`.
- Use centered positioning for the first version.
- Keep the container stable before the image loads.
- Preserve rounded corners and theme surface styling consistent with the current article reader.
- Hide the cover area entirely when an article has no cover.

This makes the cover feel like a deliberate article header instead of an optional attachment inside the article body.

## Public Article List

Article list cards should use the same cover language as the detail page. The list can remain more compact, but it should not make covers feel padded inside a frame.

Rendering rules:

- Use a stable cover container.
- Prefer `object-fit: cover` over `object-fit: contain`.
- Avoid padding around the image.
- Keep fallback behavior for broken or missing images.
- Ensure title, summary, and metadata alignment remain scannable.

The goal is for list cards to read as article cards with real covers, not as content rows with small attached thumbnails.

## Backend Editor And Preview

The backend article editor keeps the current upload and remove workflow. It should add better guidance and preview after a cover is selected or uploaded.

The cover area should show:

- The recommended `16:9` ratio and suggested pixel sizes.
- A detail-page preview using the article header treatment.
- A list-card preview showing how the cover will appear in article lists.
- A social-share preview note explaining that the same cover is used for `og:image` and `twitter:image`.
- Non-blocking warnings for low resolution or aspect ratio mismatch.

No upload action should be blocked solely because the image ratio or size is not recommended.

## Validation And Warnings

Frontend cover validation should load the selected image to inspect natural width and height.

Recommended warning behavior:

- If dimensions cannot be read, show a non-blocking warning and allow the owner to continue.
- If width or height is below `1200x675`, show a warning that the cover may look soft or pixelated.
- If aspect ratio is far from `16:9`, show a warning that the cover may be visibly cropped in the detail page, list card, or sharing previews.
- If the image matches the recommendation, show a quiet success or neutral guidance state.

The exact tolerance can be implemented as a small frontend constant so it can be tuned without searching through components.

## Component Boundaries

Keep the implementation easy to extend by separating rules from views:

- A cover policy module should own constants such as ratio, recommended dimensions, minimum dimensions, and ratio tolerance.
- A reusable cover display component or shared CSS class should own the stable `16:9` container and `object-cover` behavior.
- Backend editor preview code should consume the same constants instead of duplicating dimensions and warning text.

These boundaries allow a future version to add focal point metadata or generated derivatives without changing every rendering surface.

## Data Flow

The first version keeps the current data flow:

```text
Upload/select image -> receive cover URL -> read natural dimensions in frontend -> show previews and warnings -> save article with existing cover field
```

No API request shape changes are required. No database migration is required.

## Error Handling

- Missing cover: hide the detail-page cover and keep article layout stable.
- Broken cover in public list: fall back to the existing default image behavior.
- Broken cover in public detail: hide the cover or show a restrained fallback, but never break the reader surface.
- Image dimension read failure: show a non-blocking warning in the backend editor.
- Existing articles with unusual covers: continue rendering, using the same warnings only when edited.

## Testing

Frontend tests should cover:

- Cover policy classification for recommended, low-resolution, off-ratio, and unreadable images.
- Public article detail rendering with and without a cover.
- Public article list cover rendering using the unified stable container.
- Backend editor cover warning behavior for low-resolution and off-ratio images.
- Existing SEO behavior that maps article cover to Open Graph and Twitter metadata.

Recommended verification after implementation:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
```

## Future Extensions

Possible later improvements:

- Add focal point metadata for better automatic cropping.
- Generate dedicated list, detail, and social-share derivatives.
- Add `srcset` and modern image formats.
- Add a manual cropper if cover authoring becomes a frequent pain point.

These are intentionally left out of the first version so the immediate product improvement remains small, understandable, and easy to review.
