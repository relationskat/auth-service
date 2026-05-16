package store

const (
	createUserQuery = `
        INSERT INTO users (id, last_name, first_name, middle_name, email, password_hash)
        VALUES ($1, $2, $3, $4, $5, $6)`

	acceptEmailQuery = `
        UPDAYE users
        SET email_verified = true WHERE email = $1
    `

	loginQuery = `
        SELECT id, password_hash
        WHERE email = $1
    `

	updateUserQuery = `                                                                            
        UPDATE users                                                                             
        SET last_name   = COALESCE($2, last_name),                                               
            first_name  = COALESCE($3, first_name),                                              
            middle_name = COALESCE($4, middle_name),                                             
            updated_at  = CURRENT_TIMESTAMP                                                      
        WHERE id = $1 AND deleted_at IS NULL                                                     
        RETURNING id, last_name, first_name, middle_name`

	getPasswordHashByIDQuery = `
      SELECT password_hash
      FROM users
      WHERE id = $1 AND deleted_at IS NULL`

	updatePasswordByIDQuery = `
      UPDATE users
      SET password_hash = $2, updated_at = CURRENT_TIMESTAMP
      WHERE id = $1 AND deleted_at IS NULL`

	updatePasswordByEmailQuery = `
      UPDATE users
      SET password_hash = $2, updated_at = CURRENT_TIMESTAMP
      WHERE email = $1 AND deleted_at IS NULL`

	deleteUserQuery = `
      UPDATE users
      SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
      WHERE id = $1 AND deleted_at IS NULL`

	getEmailStatusQuery = `
      SELECT first_name, email_confirmed
      FROM users
      WHERE email = $1 AND deleted_at IS NULL`
	restoreUserByEmailQuery = `
      UPDATE users
      SET deleted_at = NULL, updated_at = CURRENT_TIMESTAMP
      WHERE email = $1 AND deleted_at IS NOT NULL`

	restoreUserByIDQuery = `
      UPDATE users
      SET deleted_at = NULL, updated_at = CURRENT_TIMESTAMP
      WHERE id = $1 AND deleted_at IS NOT NULL`

	getUserByIDQuery = `
      SELECT id, last_name, first_name, middle_name, email, email_confirmed
      FROM users
      WHERE id = $1 AND deleted_at IS NULL`
)
