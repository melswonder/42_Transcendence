-- Create index "ux_users_email" to table: "users"
CREATE UNIQUE INDEX "ux_users_email" ON "users" ("email") WHERE ((email IS NOT NULL) AND (anonymized_at IS NULL));
-- Create index "ux_users_handle" to table: "users"
CREATE UNIQUE INDEX "ux_users_handle" ON "users" ("handle") WHERE (anonymized_at IS NULL);
