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
			name:      "nil actor returns ErrUnauthorized",
			actor:     nil,
			requested: &branchA,
			want:      nil,
			wantErr:   apperror.ErrUnauthorized,
		},
		{
			name: "branch_admin with nil BranchID returns ErrForbidden",
			actor: &middleware.JWTClaims{
				Role:     "branch_admin",
				BranchID: nil,
			},
			requested: &branchA,
			want:      nil,
			wantErr:   apperror.ErrForbidden,
		},
		{
			name: "branch_admin with nil requested branch returns own BranchID",
			actor: &middleware.JWTClaims{
				Role:     "branch_admin",
				BranchID: &branchA,
			},
			requested: nil,
			want:      &branchA,
			wantErr:   nil,
		},
		{
			name: "branch_admin with matching requested branch returns own BranchID",
			actor: &middleware.JWTClaims{
				Role:     "branch_admin",
				BranchID: &branchA,
			},
			requested: &branchA,
			want:      &branchA,
			wantErr:   nil,
		},
		{
			name: "branch_admin requesting different branch returns ErrForbidden",
			actor: &middleware.JWTClaims{
				Role:     "branch_admin",
				BranchID: &branchA,
			},
			requested: &branchB,
			want:      nil,
			wantErr:   apperror.ErrForbidden,
		},
		{
			name: "owner with nil requested branch returns nil (network wide)",
			actor: &middleware.JWTClaims{
				Role: "owner",
			},
			requested: nil,
			want:      nil,
			wantErr:   nil,
		},
		{
			name: "owner requesting specific branch returns requested branch ID",
			actor: &middleware.JWTClaims{
				Role: "owner",
			},
			requested: &branchB,
			want:      &branchB,
			wantErr:   nil,
		},
		{
			name: "customer actor requesting specific branch returns requested branch ID",
			actor: &middleware.JWTClaims{
				Role: "customer",
			},
			requested: &branchA,
			want:      &branchA,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveBranchScope(tt.actor, tt.requested)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ResolveBranchScope() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ResolveBranchScope() unexpected error: %v", err)
			}

			if tt.want == nil {
				if got != nil {
					t.Errorf("ResolveBranchScope() got = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("ResolveBranchScope() got nil, want %v", *tt.want)
				} else if *got != *tt.want {
					t.Errorf("ResolveBranchScope() got = %v, want %v", *got, *tt.want)
				}
			}
		})
	}
}
