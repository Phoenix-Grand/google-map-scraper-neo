# Changelog

This file records Neo releases. The upstream base is Google Maps Scraper 1.18.1; Neo changes increment the `neo.N` suffix. Dates use UTC.

## [1.18.1-neo.2](https://github.com/Phoenix-Grand/google-map-scraper-neo/releases/tag/v1.18.1-neo.2) — 2026-10-07

### Added

- US ZIP and ZIP+4 targeting in Web UI Location Settings and the job API. ZIP+4 resolves through its first five digits; leading zeros are preserved. The resolved city, state, and ZIP appear in the job list.
- An optional radius in miles, applied to exported results in both normal and fast modes. Distances use the ZIP's approximate center or manual coordinates. Results without usable coordinates are excluded when a radius is set.
- Business details alongside the map: name, website, physical address, available phone numbers, and opening hours. Results without map coordinates remain visible when no radius is set.
- A five-column Business CSV download, available through the Web UI and `/api/v1/jobs/{id}/download?format=business`. Missing information is blank, and hours use readable day-by-day text.
- Extraction of multiple published phone numbers in both scrape modes. Web UI full CSV exports append a `phone_numbers` JSON array; command-line CSV columns remain unchanged for resume compatibility.
- A prioritized [roadmap](TODO.md) for deeper business research, the Web UI, job controls, and search coverage.

### Fixed

- Duplicate business cards when HTMX evaluates a details fragment more than once.
- Business details no longer depend on loading the optional map library. Website and map links accept HTTP/HTTPS URLs and business text is rendered as text.
- Invalid ZIPs, incomplete coordinates, and invalid radii are rejected before a job is created. Unknown ZIPs return 422; ZIP-provider failures return 503.

### Release and compatibility

- Git tag: `v1.18.1-neo.2`; public container tags: `ghcr.io/phoenix-grand/google-map-scraper-neo:1.18.1-neo.2` and `latest-neo`.
- Native Linux AMD64 and ARM64 builds include Chromium. OCI labels and image-index annotations identify the version, source revision, repository, documentation, and MIT license.
- Existing SQLite job data and CSV files remain readable without a migration. The default API download remains the full CSV; the legacy API `radius` field remains measured in meters for fast mode.
- ZIP lookup uses outbound HTTPS to Zippopotam.us. Radius targeting measures straight-line distance from a center, not ZIP boundary polygons. Search depth and runtime still limit coverage.

### Validation

- Full Go tests with race detection, lint, static checks, and Web UI tests passed before release.
- Live normal-mode search: ZIP 80202, one-mile radius, 14 results; maximum verified distance 0.965 miles.
- Live fast-mode search: ZIP+4 80202-1234, half-mile radius, 20 results; maximum verified distance 0.461 miles.
- Browser checks covered ZIP lookup errors, job creation, legacy results, details reopening without duplicates, and the five-column CSV download.

## [1.18.1-neo.1](https://github.com/Phoenix-Grand/google-map-scraper-neo/releases/tag/v1.18.1-neo.1) — 2026-10-06

- Established the Neo fork from upstream 1.18.1.
- Fixed Web UI job-table updates and pagination by preserving HTMX table fragments.
- Served HTMX locally and improved form error feedback.
- Recorded failed jobs when output creation, scraper setup, or scraping fails.
- Added a local Docker Compose setup with persistent Web UI data.
- Published public Chromium-equipped Linux AMD64 and ARM64 containers, with versioned tags and `latest-neo`.

[Compare Neo releases](https://github.com/Phoenix-Grand/google-map-scraper-neo/compare/v1.18.1-neo.1...v1.18.1-neo.2)
