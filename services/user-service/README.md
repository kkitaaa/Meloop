# User administration

All administrator routes require a valid `Authorization: Bearer <session-token>` and the authenticated account's `rol` must be `ADMIN`. The service verifies the token through `AUTH_SERVICE_URL` and checks the role in PostgreSQL; client-supplied user IDs are not trusted.

Apply `supabase/migrations/20261004210000_admin_controls.sql` before using these routes. Promote the first administrator through a trusted database session:

```sql
UPDATE usuario SET rol = 'ADMIN' WHERE correo = 'admin@example.com';
```

Available through the API Gateway:

- `GET /admin/users` lists safe account fields, suspension status, role, and moderation capability.
- `PATCH /admin/users/:id/suspension` with `{"suspendido":true}` suspends an account; `false` reactivates it.
- `PATCH /admin/users/:id/moderator` with `{"habilitado":true}` grants moderation capability; `false` removes it.

Administrators cannot change their own suspension or moderation capability. Suspended users cannot log in, and their existing session is removed on its next validation.