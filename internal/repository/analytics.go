package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/dto"
	"github.com/google/uuid"
)

// Revenue only counts orders that were actually fulfilled or are on their way
// to being fulfilled. Counting cancelled and rejected orders as revenue — as
// the previous SUM(grand_total) over every row did — overstates takings.
const revenueStatuses = `('accepted','preparing','ready','on_the_way','picked_up','delivered','completed')`

// dayBounds returns the local-midnight boundaries for "today" and the
// equivalent window yesterday, used for the day-over-day comparison.
func dayBounds(loc *time.Location, now time.Time) (todayStart, tomorrowStart, yesterdayStart time.Time) {
	local := now.In(loc)
	todayStart = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	return todayStart, todayStart.AddDate(0, 0, 1), todayStart.AddDate(0, 0, -1)
}

func pctDelta(current, previous int) *float64 {
	if previous == 0 {
		// No baseline to compare against: report "no data" rather than a
		// meaningless +100%.
		return nil
	}
	d := (float64(current) - float64(previous)) / float64(previous) * 100.0
	return &d
}

// GetDashboardStats computes today's figures for one branch, or for the whole
// network when branchID is nil.
func (r *Repository) GetDashboardStats(ctx context.Context, branchID *uuid.UUID, loc *time.Location) (*dto.DashboardStats, error) {
	now := time.Now()
	todayStart, tomorrowStart, yesterdayStart := dayBounds(loc, now)

	summary := fmt.Sprintf(`
		SELECT
			COUNT(*) FILTER (WHERE created_at >= $2 AND created_at < $3),
			COALESCE(SUM(grand_total) FILTER (WHERE created_at >= $2 AND created_at < $3 AND status IN %[1]s), 0),
			COUNT(*) FILTER (WHERE created_at >= $2 AND created_at < $3 AND status = 'pending'),
			COUNT(*) FILTER (WHERE created_at >= $2 AND created_at < $3 AND status IN ('delivered','completed')),
			COUNT(*) FILTER (WHERE created_at >= $2 AND created_at < $3 AND status IN ('cancelled','rejected')),
			COUNT(*) FILTER (WHERE created_at >= $4 AND created_at < $2),
			COALESCE(SUM(grand_total) FILTER (WHERE created_at >= $4 AND created_at < $2 AND status IN %[1]s), 0)
		FROM orders
		WHERE ($1::uuid IS NULL OR branch_id = $1)
	`, revenueStatuses)

	var stats dto.DashboardStats
	var prevOrders, prevRevenue int
	err := r.db.QueryRow(ctx, summary, branchID, todayStart, tomorrowStart, yesterdayStart).Scan(
		&stats.TotalOrders, &stats.TotalRevenue, &stats.PendingOrders,
		&stats.CompletedOrders, &stats.CancelledOrders,
		&prevOrders, &prevRevenue,
	)
	if err != nil {
		return nil, err
	}

	if stats.TotalOrders > 0 {
		stats.AvgOrder = stats.TotalRevenue / stats.TotalOrders
	}
	stats.OrdersDeltaPct = pctDelta(stats.TotalOrders, prevOrders)
	stats.RevenueDeltaPct = pctDelta(stats.TotalRevenue, prevRevenue)

	hourly, err := r.hourlySales(ctx, branchID, todayStart, tomorrowStart, loc)
	if err != nil {
		return nil, err
	}
	stats.Hourly = hourly

	if branchID == nil {
		branches, err := r.branchBreakdown(ctx, todayStart, tomorrowStart)
		if err != nil {
			return nil, err
		}
		stats.Branches = branches
	}

	return &stats, nil
}

// hourlySales returns one point per trading hour of today. Hours with no
// orders are emitted as zeroes so the chart has a continuous x-axis instead of
// collapsing gaps.
func (r *Repository) hourlySales(ctx context.Context, branchID *uuid.UUID, from, to time.Time, loc *time.Location) ([]dto.HourlySalesPoint, error) {
	query := fmt.Sprintf(`
		SELECT EXTRACT(HOUR FROM created_at AT TIME ZONE $4)::int AS hour_of_day,
		       COUNT(*)::int,
		       COALESCE(SUM(grand_total) FILTER (WHERE status IN %s), 0)::int
		FROM orders
		WHERE created_at >= $2 AND created_at < $3
		  AND ($1::uuid IS NULL OR branch_id = $1)
		GROUP BY hour_of_day
	`, revenueStatuses)

	rows, err := r.db.Query(ctx, query, branchID, from, to, loc.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type bucket struct{ orders, revenue int }
	byHour := make(map[int]bucket, 24)
	minHour, maxHour := 23, 0

	for rows.Next() {
		var h, orders, revenue int
		if err := rows.Scan(&h, &orders, &revenue); err != nil {
			return nil, err
		}
		byHour[h] = bucket{orders: orders, revenue: revenue}
		if h < minHour {
			minHour = h
		}
		if h > maxHour {
			maxHour = h
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Always span at least the typical trading window so an early-morning
	// dashboard is not a single lonely bar.
	if minHour > 8 {
		minHour = 8
	}
	nowHour := time.Now().In(loc).Hour()
	if maxHour < nowHour {
		maxHour = nowHour
	}
	if maxHour < minHour {
		maxHour = minHour
	}

	points := make([]dto.HourlySalesPoint, 0, maxHour-minHour+1)
	for h := minHour; h <= maxHour; h++ {
		b := byHour[h]
		points = append(points, dto.HourlySalesPoint{
			Hour:    fmt.Sprintf("%02d:00", h),
			Orders:  b.orders,
			Revenue: b.revenue,
		})
	}
	return points, nil
}

// branchBreakdown gives the owner a per-outlet split for today.
func (r *Repository) branchBreakdown(ctx context.Context, from, to time.Time) ([]dto.BranchSalesPoint, error) {
	query := fmt.Sprintf(`
		SELECT b.id, b.name,
		       COUNT(o.id)::int,
		       COALESCE(SUM(o.grand_total) FILTER (WHERE o.status IN %s), 0)::int
		FROM branches b
		LEFT JOIN orders o ON o.branch_id = b.id AND o.created_at >= $1 AND o.created_at < $2
		GROUP BY b.id, b.name
		ORDER BY b.name ASC
	`, revenueStatuses)

	rows, err := r.db.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]dto.BranchSalesPoint, 0)
	for rows.Next() {
		var p dto.BranchSalesPoint
		if err := rows.Scan(&p.BranchID, &p.BranchName, &p.Orders, &p.Revenue); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
