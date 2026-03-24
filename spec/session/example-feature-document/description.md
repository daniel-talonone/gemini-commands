**Note:** This is a fake feature document used for testing the `/session` commands in the Gemini CLI.

### Problem Description

As an account manager using the Admin Dashboard, I want to see a customer's membership status displayed on their profile page. This will allow me to quickly understand their account tier and provide appropriate support.

### Acceptance Criteria

1.  When a customer's profile is viewed in the Admin Dashboard, a request is made to the backend to fetch their membership data from the External API.
2.  The customer's current membership tier (e.g., "Gold", "Silver", "Bronze") is displayed clearly in a dedicated section.
3.  If the customer is not part of the membership program, a message like "Not a member" is displayed.
4.  If the backend service is unavailable or there's an error, an appropriate error message is shown.

### Technical Notes

*   The primary backend logic will be handled by the `user-service`.
*   A new endpoint, `/api/users/{id}/membership`, will be needed to fetch membership information.
*   The existing mock server can be used to mock the External API for testing this new endpoint.
