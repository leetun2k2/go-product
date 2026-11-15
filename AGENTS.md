# Agent Instructions

## Product and Priority

This is the Product Catalog Service (`go-product`), providing a canonical data
source for products, plans, prices, and quota policies across multiple websites.
Prioritize data integrity, contracts, read model consistency, and performance.
Build the technical core strictly according to the specifications:

- [Product Catalog Service Spec](specs/2026-10-06-product-catalog-service.md)
- [Database Schema Spec](specs/2026-10-06-database-schema.md)

The checked-in migrations in `./migrations/` define the storage schema.

## Working Rules

- Act as a Principal Software Engineer. Use concise English for code and docs.
- Inspect schema, callers, and consumers before changing a contract.
- Follow the Uber Go Style Guide. Use simple, idiomatic Go and explicit branches.
- Prefer guard clauses, named conditions, and named functions for callbacks.
- Avoid generic frameworks, factory patterns, and unrelated refactors.
- Preserve user edits and untracked files. Never reset the workspace.
- Do not use GitNexus, write unit tests, or verify with screenshots.
- Use small Mermaid diagrams for architecture and flows.

## Backend Boundaries

- All entity primary keys use application-generated UUIDv7.
- Keep timestamps in UTC using `TIMESTAMPTZ`.
- Minor currency units (`BIGINT amount`) ensure absolute pricing precision.
- Free offers use `billing_cycle = 'lifetime'` and `amount = 0`.
- Paid offers use one of `daily`, `weekly`, `monthly`, `yearly`, `lifetime` with `amount > 0`.
- Plan entitlements and quota rules are strictly immutable once a plan is `active`.
- Plan activation requires at least one active offer and each active offer requires at least one active price.
- Inactive plans remain queryable internally by ID to service existing subscriptions.
- Quota policies define limits and `anchor_type` (`FIRST_USE` or `SUBSCRIPTION_DATE`); no runtime counter tracking.
- The `prices` table serves as the denormalized read model with `catalog_visible` and `catalog_data` (JSONB snapshot).
- Rebuild affected `prices` read model rows atomically within the same transaction upon any change to product, plan, offer, price, or quota rules.
- Public catalog lookups read visible price rows in a single query by `product_id` and `currency`.

## Delivery and Checks

Agy orchestrates, divides tasks, delegates implementation to Claude, reviews
results, and records phase evidence. Claude edits only its assigned scope.
Report failed CLI runs and checks with their actual status. Never claim an
agent ran, a phase passed, or a deployment succeeded without evidence.

For Go changes, format changed files, run `go build ./...` and `go vet ./...`,
then exercise relevant HTTP/SQL flows on a disposable database. No new unit
tests. Report checks that could not run. For docs, check Markdown and links.
