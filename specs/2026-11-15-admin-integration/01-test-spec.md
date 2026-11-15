# Functional test specification: admin product integration

Status: planned, not executed. Scope and proposed contracts: [00-plan.md](00-plan.md).

## Preconditions

Use an authorized disposable stack and database, migrated into the product schema,
with real adapter source/build available. Provide an admin session, a non-admin
session and valid client configuration through secure runtime settings. Use unique
resource codes per run. Execute browser calls through same-origin gateway with
cookies and CSRF; direct backend security probes use internal access. Capture
`product_id`, `plan_id`, `offer_id`, `price_id`, and `quota_id` from responses.
`P` below means `/api/product/v1/admin`. No production data or embedded secrets.

## Scenarios

| ID     | Priority | Expected behavior                                                      |
| ------ | -------- | ---------------------------------------------------------------------- |
| TC-001 | Critical | Backend and BFF enforce authentication, admin role and CSRF.           |
| TC-002 | Critical | Admin creates and reads a complete paid catalog tree.                  |
| TC-003 | Critical | Activation prerequisites, free pricing and immutable policies hold.    |
| TC-004 | High     | Commercial changes update catalog visibility and snapshots atomically. |
| TC-005 | High     | Frontend handles contract differences, failures and navigation.        |
| TC-006 | Critical | Schema and network boundaries preserve existing auth and data.         |

## API chains and assertions

### TC-001: access boundaries

Purpose: prevent unauthorized catalog access or mutation.

1. Call `GET P/products` without a session, with an expired/revoked session,
   and with a valid non-admin session: expect 401, 401, and 403 respectively.
2. Repeat read and write probes against all three backend admin aliases directly
   on the internal network; caller-supplied admin/identity headers must not help.
3. As admin, attempt `POST P/products` without valid CSRF or with a foreign
   origin: adapter denies and no product row is created. Valid CSRF succeeds.
4. Remove admin role or revoke session; next request must be denied. Restore
   the fixture using a separate authorized admin.
5. Make auth unavailable or return malformed access data in a disposable setup:
   product responds with safe 503 and no mutation. Restore dependency.

Cleanup: restore roles/dependencies; remove only disposable fixtures.

### TC-002: full catalog management

Purpose: prove data passes through browser, gateway, adapter and service.

1. `POST P/products` with unique `code`, `name`, `description`; capture product.
2. `POST P/products/{product_id}/plans` with code, name and version 1; capture plan.
3. `POST P/plans/{plan_id}/quotas` with `rule_key=requests`, `limit_amount=1000`,
   `window_duration_seconds=10800`, `anchor_type=FIRST_USE`; capture quota.
4. `POST P/plans/{plan_id}/offers` with unique code, `offer_type=paid`,
   `billing_cycle=monthly`, `sort_order=0`; capture offer.
5. `POST P/offers/{offer_id}/prices` with `currency=USD`, `amount=1000`,
   `status=active`; capture price. Activate offer and then plan with their
   `POST .../{id}/activate` endpoints.
6. Read each resource and list, refresh the page, and open its direct URL.

Assert create responses are 201, successful reads/lifecycle responses are 200,
IDs remain consistent, lists unwrap their resource keys, values persist after
refresh, and visible catalog snapshots match the same product/plan/offer/quota.
Read public catalog internally at `/api/product/v1/catalog` using `product_id`
and `currency=USD`. Ensure no ingress route was opened for this lookup.

Cleanup: retain IDs for subsequent cases; dispose the isolated database after
all dependent scenarios, because product/plan/offer delete APIs do not exist.

### TC-003: invariants and policy replacement

Purpose: prevent invalid activation and modification of existing entitlements.

1. On a fresh draft plan activate without offers, then with active offer lacking
   active price: expect 422 `invalid_plan_activation`; state remains draft.
2. Create a free lifetime offer with an active zero price; activation succeeds.
   Free/non-lifetime and free/nonzero-price writes fail; paid/zero price fails.
3. Edit/add/delete quotas on a draft plan: supported operations succeed.
4. Activate it; plan metadata and quota mutations return 409 `plan_immutable`.
   Deactivate and repeat: immutability remains. Internal plan-by-ID lookup still
   succeeds. New entitlement definitions require a new plan/version.
5. Race quota update with activation using separate connections: either quota
   change commits before activation and appears in its snapshot, or it is denied
   after activation. No committed active plan may have a later policy edit.
6. Concurrently change an affected price/offer and activate a plan: ensure the
   final state and read model satisfy the agreed activation/visibility rules.

Cleanup: isolated fixture disposal. Do not deactivate production plans for tests.
Expected durable post-deactivation behavior is proposed in the plan and requires
T1; current code is known not to enforce it.

### TC-004: commercial lifecycle and read model

Purpose: verify supported price/offer changes without replacing entitlements.

Using TC-002 IDs, upsert another currency and update USD amount; read admin
price and internal catalog immediately after each successful write. Deactivate
price via `PATCH P/prices/{price_id}` and offer via its deactivate endpoint:
corresponding catalog entries disappear while internal plan reads still work.
Reactivate supported commercial resources and verify visibility returns only
when every ancestor/status condition is satisfied. Update product metadata and
verify snapshots change. Force a write failure in an isolated database session;
assert source change and snapshot do not partially commit.

Cleanup: restore fixture states if needed by subsequent checks.

### TC-005: frontend behavior and regressions

Purpose: verify operator experience and avoid silent data corruption.

Open catalog direct links, navigate back/forward, and switch products/plans while
requests are pending. Assert stale results cannot overwrite the current resource.
Exercise empty, loading, 404, denied, 409, 422, timeout and safe 503 states. Failed
forms retain input; repeat clicks do not duplicate writes; errors are translated
by code with a safe fallback. No unsupported delete or pagination controls.

Try negative/fractional/unsafe integer amounts and quota limits, including unsafe
integer responses; prevent rounded display/edit/save. Verify zero-price free
flows and currency minor-unit labels. Confirm keyboard focus/dialog management,
visible labels, narrow/wide layouts, themes and English fallback without screenshots.
Verify IAM pages, profile, login, refresh and logout still work and credentials or
session tokens never appear in frontend storage or public configuration.

Cleanup: restore simulated failures and browser session as needed.

### TC-006: infrastructure and migration

Purpose: verify isolation and data preservation before release.

Validate Compose structurally without resolving real secrets; validate Nginx in
its actual image with upstream DNS available. Inspect network membership: only
frontend/gateway/BFF use global, core services and persistors remain internal.
Verify gateway callbacks still reach auth and BFF reaches auth/product. External
probes cannot access product service ports, databases, caches or operational ports;
unknown product paths do not bypass the BFF. Check service readiness independently.

Apply schema changes to a disposable copy of existing data. Compare IDs, counts,
constraints, indexes and representative read models; verify existing auth rows are
unchanged. Confirm product runtime resolves only its schema and cannot write auth
tables. Rehearse application rollback without dropping schema or volumes. For
cross-database cutover, verify copied data and connection switch with writes paused.

Cleanup: dispose only the identified test stack/data; preserve all release volumes.

## Coverage and gaps

TC-001 covers authorization and CSRF; TC-002–004 cover persistence, commercial
rules, immutability and read models; TC-005 covers UI/contracts and IAM regression;
TC-006 covers required schema/network isolation and recovery. No tests ran during
planning. Adapter internals, live migration history, deployment host and existing
runtime data remain unverified. Execute after implementation, recording actual
requests/statuses and safe database evidence rather than claiming design guarantees.
