// Command admintool performs the operator tasks that must never be exposed
// over HTTP: setting a staff password and listing staff accounts.
//
//	go run ./cmd/admintool list
//	go run ./cmd/admintool set-password owner@cafeolga.id
//
// The password is read from the terminal without echoing. It is never passed
// as an argument, where it would land in shell history and the process list.
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/database"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/dto"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

const minPasswordLength = 12

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cfg, err := config.Load()
	if err != nil {
		fail("config: %v", err)
	}

	pool, err := database.NewPool(cfg)
	if err != nil {
		fail("database: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	switch os.Args[1] {
	case "list":
		listStaff(ctx, pool)
	case "set-password":
		if len(os.Args) < 3 {
			fail("usage: admintool set-password <email-or-phone>")
		}
		setPassword(ctx, pool, os.Args[2])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  admintool list                            list staff accounts and whether each can sign in")
	fmt.Fprintln(os.Stderr, "  admintool set-password <email-or-phone>   set a staff password interactively")
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func listStaff(ctx context.Context, pool *pgxpool.Pool) {
	rows, err := pool.Query(ctx, `
		SELECT u.email, u.phone, u.name, u.role, COALESCE(b.name, '-') AS branch,
		       (u.password_hash IS NOT NULL) AS can_sign_in
		FROM users u
		LEFT JOIN branches b ON u.branch_id = b.id
		WHERE u.role IN ('branch_admin', 'owner')
		ORDER BY u.role DESC, u.name ASC
	`)
	if err != nil {
		fail("query: %v", err)
	}
	defer rows.Close()

	fmt.Printf("%-32s %-16s %-20s %-14s %-20s %s\n", "EMAIL", "PHONE", "NAME", "ROLE", "BRANCH", "CAN SIGN IN")
	for rows.Next() {
		var email *string
		var phone, name, role, branch string
		var canSignIn bool
		if err := rows.Scan(&email, &phone, &name, &role, &branch, &canSignIn); err != nil {
			fail("scan: %v", err)
		}
		e := "(none)"
		if email != nil {
			e = *email
		}
		status := "no - locked"
		if canSignIn {
			status = "yes"
		}
		fmt.Printf("%-32s %-16s %-20s %-14s %-20s %s\n", e, phone, name, role, branch, status)
	}
	if err := rows.Err(); err != nil {
		fail("rows: %v", err)
	}
}

func setPassword(ctx context.Context, pool *pgxpool.Pool, identifier string) {
	identifier = strings.TrimSpace(identifier)

	var (
		userID, name, role string
		clause             string
		arg                any
	)
	if strings.Contains(identifier, "@") {
		clause = "LOWER(email) = LOWER($1)"
		arg = identifier
	} else {
		normalized, err := dto.NormalizeIndonesianPhone(identifier)
		if err != nil {
			fail("%q is not a valid email address or Indonesian phone number", identifier)
		}
		clause = "phone = $1"
		arg = normalized
	}

	err := pool.QueryRow(ctx,
		`SELECT id::text, name, role FROM users WHERE `+clause, arg,
	).Scan(&userID, &name, &role)
	if err != nil {
		fail("no user matches %q (%v)", identifier, err)
	}
	if role != "branch_admin" && role != "owner" {
		fail("%s is a %s account; only staff accounts use a password", identifier, role)
	}

	fmt.Fprintf(os.Stderr, "Setting password for %s (%s, %s)\n", name, identifier, role)

	pw, err := readPassword("New password: ")
	if err != nil {
		fail("reading password: %v", err)
	}
	if err := validatePassword(pw); err != nil {
		fail("%v", err)
	}
	confirm, err := readPassword("Confirm password: ")
	if err != nil {
		fail("reading confirmation: %v", err)
	}
	if pw != confirm {
		fail("passwords do not match")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		fail("hashing: %v", err)
	}

	if _, err := pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2::uuid`,
		string(hash), userID,
	); err != nil {
		fail("update: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Password updated for %s.\n", identifier)
}

// stdin is shared across reads: a bufio.Reader created per call buffers ahead
// and swallows the next line, so the confirmation prompt would hit EOF.
var stdin = bufio.NewReader(os.Stdin)

func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		return string(b), err
	}

	// Non-interactive (CI, piped input): read a single line.
	line, err := stdin.ReadString('\n')
	line = strings.TrimRight(line, "\r\n")
	if err == io.EOF && line != "" {
		err = nil
	}
	return line, err
}

// validatePassword enforces a length and character-class floor. Staff accounts
// can see every customer's name, phone and address, so a weak password here is
// a data-protection problem, not just an account problem.
func validatePassword(pw string) error {
	if len([]rune(pw)) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSymbol = true
		}
	}

	classes := 0
	for _, ok := range []bool{hasUpper, hasLower, hasDigit, hasSymbol} {
		if ok {
			classes++
		}
	}
	if classes < 3 {
		return errors.New("password must mix at least three of: uppercase, lowercase, digits, symbols")
	}
	return nil
}
