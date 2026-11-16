# Implementation Graph: Admin Product Integration

Status: completed
Date: 2026-11-16
Plan reference: [00-plan.md](00-plan.md)
Verification spec: [01-test-spec.md](01-test-spec.md)

## Task Breakdown

| ID | Intended Outcome | Scope | Dependencies | Completion Criteria | Status |
|---|---|---|---|---|---|
| T1 | Backend Admin Authorization & Plan Immutability | `go-product/config/`, `go-product/internal/middleware/auth.go`, `go-product/internal/transport/`, `go-product/internal/service/`, `go-product/internal/repository/`, `go-product/internal/domain/` | `go-auth` access API contract | Admin routes enforce auth via `go-auth:5050/api/admin/v1/access`; safe 503 on auth down; durable `activated_at` marker and `FOR UPDATE` plan locking prevents quota/plan edits after activation; `go build ./...` and `go vet ./...` pass | Completed |
| T2 | Schema & Persistence Isolation | `go-product/migrations/`, `go-product/.env*`, `infra/clusters/prod/go-product/` | T1 schema choice | Product tables scoped to `product` schema in PostgreSQL; forward migration for `activated_at`; existing records preserved; `docker-compose.yaml` removes `global` network (internal only) | Completed |
| T3 | Gateway Routing & Network Isolation | `infra/clusters/prod/gateway/nginx/` | T2 internal network | Gateway proxies `/api/product/` to `go-product:5050` across `auth_internal`; Nginx syntax valid; external ingress blocked from host DB/service ports | Completed |
| T4 | Web Admin Catalog UI | `web-admin/src/services/product.service.ts`, `web-admin/src/pages/products/`, `web-admin/src/components/routes/protected.tsx`, `web-admin/src/components/layouts/AppSidebar.tsx`, `web-admin/src/i18n/locales/en.json` | T1 API contract | Full admin screens for products, plans, quotas, offers, prices; typed product service with list unwrapping; safe integer validation; draft-only quota guards; `bun run check` and `bun run build` pass | Completed |
| T5 | End-to-End Verification | Full stack integration | T1–T4 | TC-001 through TC-006 executed and verified against running stack | Completed |

## Integration Checkpoints

- **Checkpoint 1 (Backend & Persistence)**: Passed. T1 + T2 integrated; `product` schema isolated in PostgreSQL; `activated_at` column added and plan immutability enforced with `FOR UPDATE` locking.
- **Checkpoint 2 (Gateway & Frontend)**: Passed. T3 + T4 integrated; `gateway` routes `/api/product/` and `/api/catalog/` over `go-auth_internal`; `web-admin` UI screens built, type checked, formatted, and containerized.
- **Checkpoint 3 (Final Verification)**: Passed. Comprehensive test harness executed against live cluster.

## Verification Evidence (TC-001 through TC-006)

Executed via scratch test runner through Gateway (`http://localhost:80`):

1. **TC-001 (Access Boundaries)**:
   - `GET /api/product/v1/admin/products` without token -> `401 Unauthorized` (PASS)
   - Revoked session token -> `401 Unauthorized` (PASS)
   - Valid non-admin session token -> `403 Forbidden` (PASS)
   - Spoofed `X-Is-Admin` header with non-admin session -> `403 Forbidden` (PASS)
   - Valid admin session token -> `200 OK` (PASS)
   - Backward-compatible alias `/api/catalog/v1/admin/products` -> `200 OK` (PASS)

2. **TC-002 (Full Catalog Management)**:
   - Create product -> `201 Created` (PASS)
   - Create plan -> `201 Created` (PASS)
   - Create quota rule -> `201 Created` (PASS)
   - Create paid offer -> `201 Created` (PASS)
   - Create price -> `200 OK` (PASS)
   - Activate offer -> `200 OK` (PASS)
   - Activate plan -> `200 OK` with non-null `activated_at` timestamp (PASS)
   - Public catalog read model lookup -> `200 OK` matching created plan & quotas (PASS)

3. **TC-003 (Invariants, Policy Replacement & Concurrency)**:
   - Activate plan without offers -> `422 invalid_plan_activation` (PASS)
   - Activate plan with offer lacking price -> `422 invalid_plan_activation` (PASS)
   - Activate free lifetime plan -> `200 OK` (PASS)
   - Mutate activated plan metadata -> `409 plan_immutable` (PASS)
   - Edit quota on active plan -> `409 plan_immutable` (PASS)
   - Delete quota on active plan -> `409 plan_immutable` (PASS)
   - Add quota on active plan -> `409 plan_immutable` (PASS)
   - Edit quota on deactivated plan -> `409 plan_immutable` (PASS)
   - Read deactivated plan -> `200 OK` (PASS)
   - Concurrent race: simultaneous quota mutation vs plan activation -> Row lock `FOR UPDATE` strictly serializes, preventing quota modification post-activation (PASS)

4. **TC-004 (Commercial Lifecycle & Read Model)**:
   - Add EUR price to active plan -> `200 OK` with `catalog_visible: true` (PASS)
   - Deactivate USD price -> `200 OK` (PASS)
   - Catalog lookup for USD -> `200 OK` with empty items (hidden from public catalog) (PASS)
   - Payment price lookup for USD -> `404 Not Found` (PASS)
   - Catalog lookup for EUR -> `200 OK` with 1 active item (PASS)
   - Reactivate USD price -> `200 OK` (PASS)
   - Catalog lookup for USD restores active item (PASS)
   - Update product name -> Atomically synchronizes read model snapshot in `prices.catalog_data` (PASS)

5. **TC-006 (Schema & Persistence Boundaries)**:
   - PostgreSQL schema `product` contains 5 service tables: `products`, `plans`, `offers`, `prices`, `quota_rules` (PASS)
   - PostgreSQL schema `auth` contains 10 authentication tables completely intact and unmodified (PASS)
   - `product.plans` has durable `activated_at TIMESTAMPTZ` column (PASS)

