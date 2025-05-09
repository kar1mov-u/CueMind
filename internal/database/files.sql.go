// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: files.sql

package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const compeleteFileDetails = `-- name: CompeleteFileDetails :exec
UPDATE files SET file_name=$1,format =$2 WHERE id=$3 and collection_id=$4 and user_id=$5
`

type CompeleteFileDetailsParams struct {
	FileName     sql.NullString
	Format       sql.NullString
	ID           uuid.UUID
	CollectionID uuid.UUID
	UserID       uuid.UUID
}

func (q *Queries) CompeleteFileDetails(ctx context.Context, arg CompeleteFileDetailsParams) error {
	_, err := q.db.ExecContext(ctx, compeleteFileDetails,
		arg.FileName,
		arg.Format,
		arg.ID,
		arg.CollectionID,
		arg.UserID,
	)
	return err
}

const craeteFileEntry = `-- name: CraeteFileEntry :one
INSERT INTO files(
    collection_id, user_id
) VALUES (
    $1, $2
) 
RETURNING id
`

type CraeteFileEntryParams struct {
	CollectionID uuid.UUID
	UserID       uuid.UUID
}

func (q *Queries) CraeteFileEntry(ctx context.Context, arg CraeteFileEntryParams) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, craeteFileEntry, arg.CollectionID, arg.UserID)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const deleteAllFiles = `-- name: DeleteAllFiles :exec
DELETE FROM files WHERE collection_id=$1 and user_id=$2
`

type DeleteAllFilesParams struct {
	CollectionID uuid.UUID
	UserID       uuid.UUID
}

func (q *Queries) DeleteAllFiles(ctx context.Context, arg DeleteAllFilesParams) error {
	_, err := q.db.ExecContext(ctx, deleteAllFiles, arg.CollectionID, arg.UserID)
	return err
}

const deleteFile = `-- name: DeleteFile :exec
DELETE FROM files WHERE id=$1
`

func (q *Queries) DeleteFile(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteFile, id)
	return err
}

const getFilesForCollection = `-- name: GetFilesForCollection :many
SELECT id, collection_id, user_id, file_name, format, uploaded_at, processed FROM files WHERE collection_id=$1 and user_id = $2
`

type GetFilesForCollectionParams struct {
	CollectionID uuid.UUID
	UserID       uuid.UUID
}

func (q *Queries) GetFilesForCollection(ctx context.Context, arg GetFilesForCollectionParams) ([]File, error) {
	rows, err := q.db.QueryContext(ctx, getFilesForCollection, arg.CollectionID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []File
	for rows.Next() {
		var i File
		if err := rows.Scan(
			&i.ID,
			&i.CollectionID,
			&i.UserID,
			&i.FileName,
			&i.Format,
			&i.UploadedAt,
			&i.Processed,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const processed = `-- name: Processed :exec
UPDATE files SET processed = $1 WHERE id = $2
`

type ProcessedParams struct {
	Processed bool
	ID        uuid.UUID
}

func (q *Queries) Processed(ctx context.Context, arg ProcessedParams) error {
	_, err := q.db.ExecContext(ctx, processed, arg.Processed, arg.ID)
	return err
}

const processedCheck = `-- name: ProcessedCheck :one
SELECT processed FROM files WHERE id = $1
`

func (q *Queries) ProcessedCheck(ctx context.Context, id uuid.UUID) (bool, error) {
	row := q.db.QueryRowContext(ctx, processedCheck, id)
	var processed bool
	err := row.Scan(&processed)
	return processed, err
}
