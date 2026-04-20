# Pull Request
*This is a sample pull request description for testing purposes.*

## Description
This PR introduces the ability to view a customer's membership status directly within the Admin Dashboard. It adds a new backend endpoint to fetch the data and a corresponding UI component to display it.

### Resolves
Fixes TICKET-123

## Changes
- Adds a new endpoint `/api/users/{id}/membership` to the `user-service`.
- Mocks the External API response using the mock server.
- Creates a new UI component to display the membership tier.
- Handles cases for members, non-members, and API errors.

## How to Test
1.  Navigate to a customer profile in the Admin Dashboard.
2.  Verify that the "Membership Status" section appears.
3.  For a mocked member, you should see a tier like "Gold".
4.  For a mocked non-member, you should see "Not a member".
