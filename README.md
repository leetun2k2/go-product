# Product Catalog Service (`go-product`)

Product Catalog Service is the canonical source of truth for products, plans, prices, and quota policies across multiple websites.

## Architecture

- **Specifications:**
  - [Product Catalog Service Spec](specs/2026-10-06-product-catalog-service.md)
  - [Database Schema Specification](specs/2026-10-06-database-schema.md)
- **Database:** PostgreSQL 16+ using UUIDv7 primary keys.
- **Read Model:** Denormalized `prices` table containing `catalog_data` (JSONB) and `catalog_visible` (BOOLEAN) indexed for single-query catalog lookups.

## Key Boundaries

- **Plan Entitlement Immutability:** Once a plan transitions from `draft` to `active`, its entitlements and quota rules are immutable.
- **Unified Offers:** Free and paid plans share the same model. Free offers use `billing_cycle = 'lifetime'` and `amount = 0`.
- **Atomic Read Model Sync:** Every modification to products, plans, offers, prices, or quotas recalculates the affected `prices` read model rows inside the same transaction.

## Quickstart

```bash
# Build
go build ./...

# Run
go run ./cmd/server
```
