package service

import (
	"errors"
	"testing"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/apperror"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/middleware"
	"github.com/google/uuid"
)

func TestResolveBranchScope(t *testing.T) {
	branchA := uuid.New()
	branchB := uuid.New()

	tests := []struct {
		name      string
		actor     *middleware.JWTClaims
		requested *uuid.UUID
		want      *uuid.UUID
		wantErr   error
	}{
		{
			name:      "nil actor returns unauthorized",
			actor:     nil,
			requested: &branchA,
			want:      nil,
			wantErr:   apperror.ErrUnauthorized,
		},
		{
			name:      "customer role returns forbidden",
			actor:     &middleware.JWTClaims{Role: "customer"},
			requested: &branchA,
			want:      nil,
			wantErr:   apperror.ErrForbidden,
		},
		{
			name:      "owner with requested branch returns requested branch",
			actor:     &middleware.JWTClaims{Role: "owner"},
			requested: &branchA,
			want:      &branchA,
			wantErr:   nil,
		},
		{
			name:      "owner with nil requested returns nil",
			actor:     &middleware.JWTClaims{Role: "owner"},
			requested: nil,
			want:      nil,
			wantErr:   nil,
		},
		{
			name:      "branch_admin with missing BranchID returns forbidden",
			actor:     &middleware.JWTClaims{Role: "branch_admin", BranchID: nil},
			requested: &branchA,
			want:      nil,
			wantErr:   apperror.ErrForbidden,
		},
		{
			name:      "branch_admin requesting own branch returns own branch",
			actor:     &middleware.JWTClaims{Role: "branch_admin", BranchID: &branchA},
			requested: &branchA,
			want:      &branchA,
			wantErr:   nil,
		},
		{
			name:      "branch_admin with nil requested returns own branch",
			actor:     &middleware.JWTClaims{Role: "branch_admin", BranchID: &branchA},
			requested: nil,
			want:      &branchA,
			wantErr:   nil,
		},
		{
			name:      "branch_admin requesting different branch returns forbidden",
			actor:     &middleware.JWTClaims{Role: "branch_admin", BranchID: &branchA},
			requested: &branchB,
			want:      nil,
			wantErr:   apperror.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveBranchScope(tt.actor, tt.requested)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ResolveBranchScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil {
				if got != nil {
					t.Errorf("ResolveBranchScope() got = %v, want nil", got)
				}
			} else {
				if got == nil || *got != *tt.want {
					t.Errorf("ResolveBranchScope() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestResolveWritableBranch(t *testing.T) {
	branchA := uuid.New()
	branchB := uuid.New()

	tests := []struct {
		name      string
		actor     *middleware.JWTClaims
		requested uuid.UUID
		want      uuid.UUID
		wantErr   error
	}{
		{
			name:      "nil actor returns unauthorized",
			actor:     nil,
			requested: branchA,
			want:      uuid.Nil,
			wantErr:   apperror.ErrUnauthorized,
		},
		{
			name:      "customer role returns forbidden",
			actor:     &middleware.JWTClaims{Role: "customer"},
			requested: branchA,
			want:      uuid.Nil,
			wantErr:   apperror.ErrForbidden,
		},
		{
			name:      "owner with valid requested branch returns requested branch",
			actor:     &middleware.JWTClaims{Role: "owner"},
			requested: branchA,
			want:      branchA,
			wantErr:   nil,
		},
		{
			name:      "owner with uuid.Nil requested branch returns invalid error",
			actor:     &middleware.JWTClaims{Role: "owner"},
			requested: uuid.Nil,
			want:      uuid.Nil,
			wantErr:   apperror.Invalid("outlet wajib dipilih"),
		},
		{
			name:      "branch_admin with missing BranchID returns forbidden",
			actor:     &middleware.JWTClaims{Role: "branch_admin", BranchID: nil},
			requested: branchA,
			want:      uuid.Nil,
			wantErr:   apperror.ErrForbidden,
		},
		{
			name:      "branch_admin requesting own branch returns own branch",
			actor:     &middleware.JWTClaims{Role: "branch_admin", BranchID: &branchA},
			requested: branchA,
			want:      branchA,
			wantErr:   nil,
		},
		{
			name:      "branch_admin with uuid.Nil requested returns own branch",
			actor:     &middleware.JWTClaims{Role: "branch_admin", BranchID: &branchA},
			requested: uuid.Nil,
			want:      branchA,
			wantErr:   nil,
		},
		{
			name:      "branch_admin requesting different branch returns forbidden",
			actor:     &middleware.JWTClaims{Role: "branch_admin", BranchID: &branchA},
			requested: branchB,
			want:      uuid.Nil,
			wantErr:   apperror.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveWritableBranch(tt.actor, tt.requested)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("resolveWritableBranch() expected error %v, got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("resolveWritableBranch() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else if err != nil {
				t.Errorf("resolveWritableBranch() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("resolveWritableBranch() got = %v, want %v", got, tt.want)
			}
		})
	}
}
