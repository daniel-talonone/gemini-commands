## Strategy

Additive change. The backend already has a `user-service` with an established service/repository pattern — we extend it rather than introduce a new service. The frontend profile page fetches data via an existing API client layer; we add one new hook/component without touching unrelated profile sections.

The membership data comes from an External API, so the backend acts as a proxy with its own error boundary — the frontend never calls the External API directly.

## Pattern References

- Backend endpoint: follow `src/user-service/routes/user-preferences.ts` — same router structure, same error handling middleware, same response shape `{ data, error }`.
- Frontend data fetching: follow `src/components/profile/BillingSection.tsx` — uses the `useProfileSection(endpoint)` hook pattern, loading/error states included.
- Mock server setup: follow `mocks/handlers/billing.ts` for how to register a new External API mock.

## Constraints

- Do not call the External API directly from the frontend — all membership data must flow through `/api/users/{id}/membership`.
- Do not add membership logic to the existing `UserProfile` service method — keep it in a dedicated `getMembershipStatus(userId)` function.
- Do not render the membership section inside the existing `<ProfileHeader>` component — it gets its own `<MembershipStatus>` component so it can fail independently.
- No new shared abstractions — this is a one-off feature, not a pattern to generalize yet.

## Slice Hints

1. **Backend foundation** — add the endpoint, service function, and mock. Backend tests pass. Frontend untouched. *(safe stop)*
2. **Frontend component (disconnected)** — build `<MembershipStatus>` with hardcoded/prop-driven data, no API call yet. Component renders correctly in isolation. *(safe stop)*
3. **Wire frontend to backend** — connect the component via `useProfileSection`, handle loading/error/not-a-member states. *(safe stop)*
4. **Cleanup** — remove any temporary hardcoded values, add missing edge-case tests.
