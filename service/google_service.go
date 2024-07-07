package service

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/nilotpaul/go-downloader/types"
	"github.com/nilotpaul/go-downloader/util"
	"golang.org/x/oauth2"
)

const (
	apiEndpoint = "https://www.googleapis.com/oauth2/v3/userinfo"
)

func GetGoogleUserInfo(token *oauth2.Token, client *http.Client) (*types.GoogleUserResponse, error) {
	req, err := http.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, util.NewAppError(
			req.Response.StatusCode,
			err.Error(),
			"NewGoogleProvider, GetUserInfo error:  ",
			err,
		)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	res, err := client.Do(req)
	if err != nil {
		return nil, util.NewAppError(
			res.StatusCode,
			err.Error(),
			"NewGoogleProvider, GetUserInfo error:  ",
			err,
		)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, util.NewAppError(
			http.StatusInternalServerError,
			"failed to get the user info",
			"NewGoogleProvider, GetUserInfo error:",
			"status not ok",
		)
	}

	var userInfo types.GoogleUserResponse
	if err := util.DecodeJSON(res.Body, &userInfo); err != nil {
		return nil, util.NewAppError(
			http.StatusInternalServerError,
			"failed to decode the res body",
			"NewGoogleProvider, GetUserInfo error:  ",
			err,
		)
	}

	return &userInfo, nil
}

func CreateUserAndAccount(db *sql.DB, user *types.GoogleUserResponse, token *oauth2.Token) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer util.CommitOrRollback(tx, &err)

	const userQuery string = `
		INSERT INTO users (email)
		VALUES ($1)
		RETURNING id
	`

	var userID string
	if err = tx.QueryRow(userQuery, user.Email).Scan(&userID); err != nil {
		return "", err
	}

	const accQuery = `
		INSERT INTO google_accounts (
			user_id, 
			access_token, 
			refresh_token,
			token_type, 
			expires_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.Exec(
		accQuery,
		userID,
		token.AccessToken,
		token.RefreshToken,
		token.TokenType,
		token.Expiry,
		time.Now(),
	)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func GetUserAndAccountByEmail(db *sql.DB, email string) (*types.User, *types.GoogleAccount, error) {
	const query = `
	    SELECT 
		    u.id, u.email, 
			a.access_token, a.refresh_token, a.token_type, a.expires_at, a.created_at, a.updated_at
		FROM
		    users u
		INNER JOIN
		    google_accounts a
		ON
		    u.email = $1
	`

	var (
		u   types.User
		acc types.GoogleAccount
	)
	row := db.QueryRow(query, email)
	err := row.Scan(
		&u.UserID, &u.Email,
		&acc.AccessToken, &acc.RefreshToken, &acc.TokenType, &acc.ExpiresAt, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}

	return &u, &acc, nil
}

func GetAccountByUserID(db *sql.DB, userID string) (*types.GoogleAccount, error) {
	const query = `SELECT * FROM google_accounts WHERE user_id = $1`

	var acc types.GoogleAccount
	row := db.QueryRow(query, userID)
	err := row.Scan(
		&acc.ID,
		&acc.UserID,
		&acc.AccessToken,
		&acc.RefreshToken,
		&acc.ExpiresAt,
		&acc.CreatedAt,
		&acc.UpdatedAt,
		&acc.TokenType,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &acc, nil
}

func GetUserByEmail(db *sql.DB, email string) (*types.User, error) {
	const query = `SELECT * FROM users WHERE email = $1`

	var u types.User
	row := db.QueryRow(query, email)
	err := row.Scan(
		&u.UserID,
		&u.Email,
		&u.CreatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &u, nil
}

func GetUserByID(db *sql.DB, userID string) (*types.User, error) {
	const query = `SELECT * FROM users WHERE id = $1`

	var u types.User
	row := db.QueryRow(query, userID)
	err := row.Scan(
		&u.UserID,
		&u.Email,
		&u.CreatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &u, nil
}

func UpdateAccountByUserID(db *sql.DB, userID string, acc *types.GoogleAccount) error {
	const query = `
	    UPDATE google_accounts
		SET
		    access_token = $1,
			refresh_token = $2,
			token_type = $3,
			expires_at = $4,
			updated_at = $6
		WHERE
            user_id = $5
	`
	if _, err := db.Exec(
		query,
		acc.AccessToken,
		acc.RefreshToken,
		acc.TokenType,
		acc.ExpiresAt,
		userID,
		time.Now(),
	); err != nil {
		return err
	}

	return nil
}
