# Article List Cover UI Design

## Context

Article covers now have a canonical 16:9 policy and shared `ArticleCover` component, but list rendering is still uneven:

- The public article list uses `ArticleCover`, but the card rhythm still needs refinement around cover weight, spacing, and responsive behavior.
- The admin article list renders covers with a hand-written `<img>`, so fallback, aspect ratio, hover behavior, and error handling can drift from the public list.
- Admin list interactions are inconsistent: the title opens a preview panel, while the cover still behaves like an edit entry point.

This spec redesigns the front-end and admin article list cover UI without changing the article cover data model or upload policy.

## Goals

- Use one visual rule set for article covers across public and admin lists.
- Keep the public list as a reading-oriented card experience.
- Keep the admin list as a management-oriented list experience.
- Make cover and title clicks consistent in admin: both open preview; editing only happens through the edit button.
- Preserve information density and scanning efficiency in the admin list.
- Keep the implementation maintainable by reusing `ArticleCover` instead of duplicating image behavior.

## Non-Goals

- Do not change `article.cover` storage, upload APIs, or backend models.
- Do not introduce image cropping, focal points, or generated derivatives.
- Do not redesign article detail covers or the editor cover panel in this pass.
- Do not turn the admin article list into a full data table.

## Visual Direction

Use medium cover emphasis.

The cover should be visible enough to identify an article and add rhythm, but it must not overpower title, summary, status, metadata, or admin actions. The public list can feel more editorial, while the admin list should remain compact, utilitarian, and easy to scan.

## Public Article List

The public list should stay as reading cards:

- Desktop: horizontal card layout with cover on the left and content on the right.
- Mobile: stacked card layout with cover above content.
- Cover: stable 16:9 frame, shared fallback, object-cover cropping, subtle hover scale.
- Content hierarchy: category/read-time first, then title, summary, metadata.
- Click behavior: cover and title navigate to the article detail page.

The card should feel more intentional than a raw media object beside text: cover width, card padding, hover treatment, and content spacing should be tuned together.

## Admin Article List

The admin list should stay as management rows:

- Desktop: horizontal row layout with a compact 16:9 cover thumbnail on the left.
- Mobile: stacked layout with cover above content when horizontal space is too narrow.
- Cover: reuse `ArticleCover` with the same fallback and image behavior as the public list.
- Content hierarchy: status/category badges, title, summary, metadata, then action buttons.
- Click behavior: cover and title open the preview panel.
- Edit behavior: editing is available only through the edit button.

The cover should help identify content quickly without increasing row height unnecessarily. Admin actions must remain visually and interactively separate from preview entry points.

## Component Boundary

`ArticleCover` remains the shared primitive for cover rendering:

- It owns image source normalization, fallback display, fixed 16:9 frame, object-cover behavior, and image error handling.
- List views own layout, click target semantics, spacing, and responsive arrangement.
- The admin list should stop using a hand-written `<img>` for covers.

Avoid adding a new abstraction unless the implementation reveals repeated layout code that is harder to maintain than a small wrapper component.

## States And Accessibility

- Loading states should match the final card or row proportions closely enough to avoid layout jumps.
- Missing covers should render a deliberate fallback, not a broken-image feeling.
- Admin cover preview buttons must expose clear accessible names, such as previewing the article cover/title.
- Hover and focus styles should be visible but restrained.
- Mobile layouts must avoid squeezed thumbnails, overlapping actions, or text truncation that hides primary controls.

## Acceptance Criteria

- Public list covers keep a medium-emphasis reading-card presentation on desktop and mobile.
- Admin list covers use the shared `ArticleCover` component.
- Admin cover clicks open the preview panel, matching title clicks.
- Admin editing is reachable only through the edit button.
- Fallback, crop, hover, and aspect-ratio behavior no longer diverge between front-end and admin list covers.
- Responsive layouts remain usable at narrow widths.
- Focused tests or source-contract tests cover the shared rendering and admin click contract.

