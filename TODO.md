# Future additions

Proposed work for Google Maps Scraper Neo, ordered by priority. These features are not commitments or part of the current release. Reuse existing scraper, email, review, grid, and resume capabilities where possible.

## Next priorities

- [ ] **“Dig deeper” option.** Let users enrich a whole job or selected businesses after the initial scrape. Reuse email and extended-review collection, then visit the business's public website/contact/about pages for additional phone numbers, emails, social links, services, and descriptions. Record source URLs and collection dates, distinguish missing from verified information, and expose page/time limits before starting.
- [ ] **Web UI facelift.** Improve desktop and mobile layouts, simplify job creation, and make business details easier to scan. Add accessible navigation, clear empty/error states, a result count, and a consistent visual style. Keep advanced settings available without overwhelming the main form.
- [ ] **Searchable results.** Add filtering, sorting, and table/card/map views. Let users find businesses missing websites or phones, select rows, and export only the selected results.
- [ ] **Progress and job controls.** Show elapsed time, businesses collected, radius exclusions, retries, and useful failure messages. Add stop, duplicate, and retry actions; reuse command-line resume behavior for interrupted Web UI jobs.
- [ ] **SSO session handling.** Detect expired authentication during polling or form submission, preserve entered settings, and send users back through sign-in with a clear message. Document tested Nginx/OAuth2 Proxy configurations and keep the authentication challenge enforced.

## Search quality and coverage

- [ ] **Location preview.** Show the resolved ZIP center and radius circle before a job starts, with an option to confirm or adjust the center. Make the distinction between ZIP centers and ZIP boundaries clear.
- [ ] **Wider area coverage.** Expose existing grid-search capabilities in the Web UI to cover a large radius more thoroughly. Estimate search cells and runtime, cap job size, and deduplicate businesses across cells.
- [ ] **Multiple target areas.** Accept a list of ZIPs or saved territories, preserve the source target for every result, and deduplicate overlapping searches.
- [ ] **Postal-code providers and caching.** Cache lookups, add configurable providers and timeouts, and support other countries with an explicit country selector. Consider ZIP boundary polygons as a separate targeting mode.
- [ ] **Collection diagnostics.** Show whether a search stopped because it finished, hit a depth/time limit, encountered throttling, or could not load a page. Explain normal/fast-mode coverage differences.

## Business data and exports

- [ ] **Hours and contact normalization.** Preserve raw published values while also offering structured hours, timezone, international phone formatting, and per-field source information. Support split opening periods, closed days, and multiple phones without guessing missing details.
- [ ] **Persistent business records.** Match businesses across jobs using place IDs and other existing stable identities. Show duplicates and changes to websites, addresses, phones, or hours between collection dates.
- [ ] **More export options.** Add selected-column CSV, JSON, and spreadsheet downloads in the Web UI, plus reusable export presets. Keep existing API and command-line formats compatible.
- [ ] **Saved campaigns.** Save keywords, locations, radius, and enrichment settings for repeat runs. Add an optional scheduler and completion webhooks after job progress and retry behavior are reliable.

## Reliability and maintenance

- [ ] **Browser regression coverage.** Automate tests for ZIP errors, fractional radii, missing details, CSV downloads, map-library failures, and repeated HTMX fragment evaluation. Use deterministic fixtures rather than relying on live Google Maps for every test.
- [ ] **Container operations.** Add health/readiness checks, configurable resource limits, and documented backup/restore procedures for Web UI data.
- [ ] **Observable runs.** Add structured job logs, retention settings, and a support bundle that excludes credentials and private proxy settings.
- [ ] **Upstream maintenance.** Track upstream parser/browser changes, preserve Neo behavior during merges, and verify both container platforms before publishing each release.

Suggested first milestone: the UI facelift, searchable results, progress/stop/retry controls, and a bounded “dig deeper” flow with field provenance. Add scheduling and larger-area searches after those workflows are dependable.
