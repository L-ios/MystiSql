package postgresql

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type PostgreSQLError struct {
	Code       string
	Message    string
	Detail     string
	Constraint string
	Table      string
	Column     string
	Original   error
}

func (e *PostgreSQLError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *PostgreSQLError) Unwrap() error {
	return e.Original
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if AsPgError(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if AsPgError(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

func IsCheckViolation(err error) bool {
	var pgErr *pgconn.PgError
	if AsPgError(err, &pgErr) {
		return pgErr.Code == "23514"
	}
	return false
}

func IsNotNullViolation(err error) bool {
	var pgErr *pgconn.PgError
	if AsPgError(err, &pgErr) {
		return pgErr.Code == "23502"
	}
	return false
}

func IsSyntaxError(err error) bool {
	var pgErr *pgconn.PgError
	if AsPgError(err, &pgErr) {
		return pgErr.Code == "42601"
	}
	return false
}

func IsUndefinedTable(err error) bool {
	var pgErr *pgconn.PgError
	if AsPgError(err, &pgErr) {
		return pgErr.Code == "42P01"
	}
	return false
}

func IsUndefinedColumn(err error) bool {
	var pgErr *pgconn.PgError
	if AsPgError(err, &pgErr) {
		return pgErr.Code == "42703"
	}
	return false
}

func AsPgError(err error, pgErr **pgconn.PgError) bool {
	if err == nil {
		return false
	}

	if p, ok := err.(*pgconn.PgError); ok {
		*pgErr = p
		return true
	}

	return false
}

func WrapError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !AsPgError(err, &pgErr) {
		return err
	}

	return &PostgreSQLError{
		Code:       pgErr.Code,
		Message:    pgErr.Message,
		Detail:     pgErr.Detail,
		Constraint: pgErr.ConstraintName,
		Table:      pgErr.TableName,
		Column:     pgErr.ColumnName,
		Original:   err,
	}
}

func GetErrorDetail(err error) map[string]string {
	var pgErr *pgconn.PgError
	if !AsPgError(err, &pgErr) {
		return nil
	}

	detail := make(map[string]string)
	detail["code"] = pgErr.Code
	detail["message"] = pgErr.Message

	if pgErr.Detail != "" {
		detail["detail"] = pgErr.Detail
	}
	if pgErr.ConstraintName != "" {
		detail["constraint"] = pgErr.ConstraintName
	}
	if pgErr.TableName != "" {
		detail["table"] = pgErr.TableName
	}
	if pgErr.ColumnName != "" {
		detail["column"] = pgErr.ColumnName
	}

	return detail
}
