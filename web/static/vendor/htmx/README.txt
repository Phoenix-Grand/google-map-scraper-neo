HTMX 1.9.6 (unchanged from the original Web UI dependency).

Source: https://cdnjs.cloudflare.com/ajax/libs/htmx/1.9.6/htmx.min.js
Upstream: https://github.com/bigskysoftware/htmx/tree/v1.9.6
License: BSD 2-Clause; see LICENSE in this directory.
SHA-256: cbb723c305cf6d6315c890909815523588509e2e092a59f8cfc4a885829689d5

Served locally through Go's embedded static files so core UI controls do not
depend on CDN availability. useTemplateFragments preserves tbody elements
alongside the out-of-band pagination fragment during HTMX swaps.
