package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/dto"
	"github.com/ericanthonywu/maremereso-olga/backend/internal/model"
	"github.com/google/uuid"
)

type branchAccumulator struct {
	id         uuid.UUID
	name       string
	slug       string
	count      int
	sumOverall float64
	sumResto   float64
	countResto int
	sumApp     float64
	countApp   int
}

type menuItemAccumulator struct {
	name          string
	count         int
	sumRating     float64
	positiveCount int
	negativeCount int
	sampleReasons []string
}

type tagAccumulator struct {
	tag        string
	count      int
	isPositive bool
}

// ListOrderFeedbackAdmin fetches paginated customer feedbacks with filtering.
func (r *Repository) ListOrderFeedbackAdmin(
	ctx context.Context,
	branchID *uuid.UUID,
	rating *int,
	search string,
	limit, offset int,
) ([]model.OrderFeedbackAdminItem, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	search = strings.TrimSpace(search)
	var searchArg any
	if search != "" {
		searchArg = "%" + search + "%"
	}

	whereClauses := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if branchID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("o.branch_id = $%d", argIdx))
		args = append(args, *branchID)
		argIdx++
	}

	if rating != nil && *rating >= 1 && *rating <= 5 {
		whereClauses = append(whereClauses, fmt.Sprintf("(f.rating = $%d OR f.resto_rating = $%d OR f.app_rating = $%d)", argIdx, argIdx, argIdx))
		args = append(args, *rating)
		argIdx++
	}

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(`(
			o.order_number ILIKE $%d OR
			o.customer_name ILIKE $%d OR
			o.customer_phone ILIKE $%d OR
			COALESCE(f.comment, '') ILIKE $%d OR
			COALESCE(f.resto_reason, '') ILIKE $%d OR
			COALESCE(f.app_reason, '') ILIKE $%d OR
			COALESCE(f.items_feedback::text, '') ILIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx, argIdx, argIdx, argIdx))
		args = append(args, searchArg)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := `
		SELECT COUNT(*)
		FROM order_feedback f
		JOIN orders o ON f.order_id = o.id
		JOIN branches b ON o.branch_id = b.id
		WHERE ` + whereSQL

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			f.order_id,
			o.order_number,
			o.customer_name,
			o.customer_phone,
			o.branch_id,
			b.name AS branch_name,
			b.slug AS branch_slug,
			f.rating,
			f.resto_rating,
			f.app_rating,
			f.resto_reason,
			f.app_reason,
			f.comment,
			COALESCE(f.items_feedback, '[]'::jsonb) AS items_feedback,
			f.created_at,
			f.updated_at
		FROM order_feedback f
		JOIN orders o ON f.order_id = o.id
		JOIN branches b ON o.branch_id = b.id
		WHERE %s
		ORDER BY f.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.OrderFeedbackAdminItem, 0)
	for rows.Next() {
		var item model.OrderFeedbackAdminItem
		var itemsFeedbackRaw []byte
		if err := rows.Scan(
			&item.OrderID,
			&item.OrderNumber,
			&item.CustomerName,
			&item.CustomerPhone,
			&item.BranchID,
			&item.BranchName,
			&item.BranchSlug,
			&item.Rating,
			&item.RestoRating,
			&item.AppRating,
			&item.RestoReason,
			&item.AppReason,
			&item.Comment,
			&itemsFeedbackRaw,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if len(itemsFeedbackRaw) > 0 {
			_ = json.Unmarshal(itemsFeedbackRaw, &item.ItemsFeedback)
		}
		if item.ItemsFeedback == nil {
			item.ItemsFeedback = []model.OrderItemFeedback{}
		}
		items = append(items, item)
	}

	return items, total, nil
}

// GetFeedbackAnalytics computes aggregated review metrics, rating distributions,
// branch comparisons, and menu item performance.
func (r *Repository) GetFeedbackAnalytics(ctx context.Context, branchID *uuid.UUID) (*dto.FeedbackAnalytics, error) {
	query := `
		SELECT 
			f.order_id,
			o.branch_id,
			b.name AS branch_name,
			b.slug AS branch_slug,
			f.rating,
			f.resto_rating,
			f.app_rating,
			COALESCE(f.resto_reason, ''),
			COALESCE(f.app_reason, ''),
			COALESCE(f.comment, ''),
			COALESCE(f.items_feedback, '[]'::jsonb)
		FROM order_feedback f
		JOIN orders o ON f.order_id = o.id
		JOIN branches b ON o.branch_id = b.id
		WHERE ($1::uuid IS NULL OR o.branch_id = $1)
	`
	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalReviews int
	var sumOverall, sumResto, sumApp float64
	var countOverall, countResto, countApp int
	ratingCounts := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	branchStatsMap := make(map[uuid.UUID]*branchAccumulator)
	menuStatsMap := make(map[string]*menuItemAccumulator)
	tagCountsMap := make(map[string]*tagAccumulator)

	for rows.Next() {
		var orderID, bID uuid.UUID
		var bName, bSlug, restoReason, appReason, comment string
		var rating int
		var restoRating, appRating *int
		var itemsRaw []byte

		if err := rows.Scan(
			&orderID, &bID, &bName, &bSlug,
			&rating, &restoRating, &appRating,
			&restoReason, &appReason, &comment,
			&itemsRaw,
		); err != nil {
			return nil, err
		}

		totalReviews++

		// Overall rating
		effectiveOverall := rating
		if effectiveOverall < 1 || effectiveOverall > 5 {
			if restoRating != nil && *restoRating >= 1 && *restoRating <= 5 {
				effectiveOverall = *restoRating
			} else if appRating != nil && *appRating >= 1 && *appRating <= 5 {
				effectiveOverall = *appRating
			} else {
				effectiveOverall = 5
			}
		}
		sumOverall += float64(effectiveOverall)
		countOverall++
		ratingCounts[effectiveOverall]++

		// Resto rating
		if restoRating != nil && *restoRating >= 1 && *restoRating <= 5 {
			sumResto += float64(*restoRating)
			countResto++
		} else {
			sumResto += float64(effectiveOverall)
			countResto++
		}

		// App rating
		if appRating != nil && *appRating >= 1 && *appRating <= 5 {
			sumApp += float64(*appRating)
			countApp++
		}

		// Branch accumulator
		bAcc, exists := branchStatsMap[bID]
		if !exists {
			bAcc = &branchAccumulator{
				id:   bID,
				name: bName,
				slug: bSlug,
			}
			branchStatsMap[bID] = bAcc
		}
		bAcc.count++
		bAcc.sumOverall += float64(effectiveOverall)
		if restoRating != nil && *restoRating >= 1 && *restoRating <= 5 {
			bAcc.sumResto += float64(*restoRating)
			bAcc.countResto++
		} else {
			bAcc.sumResto += float64(effectiveOverall)
			bAcc.countResto++
		}
		if appRating != nil && *appRating >= 1 && *appRating <= 5 {
			bAcc.sumApp += float64(*appRating)
			bAcc.countApp++
		}

		// Process tags from reasons
		processReasonTags(restoReason, tagCountsMap)
		processReasonTags(appReason, tagCountsMap)

		// Menu items feedback
		if len(itemsRaw) > 0 {
			var items []model.OrderItemFeedback
			if err := json.Unmarshal(itemsRaw, &items); err == nil {
				for _, it := range items {
					if it.ItemName == "" || it.Rating < 1 || it.Rating > 5 {
						continue
					}
					mAcc, mExists := menuStatsMap[it.ItemName]
					if !mExists {
						mAcc = &menuItemAccumulator{name: it.ItemName}
						menuStatsMap[it.ItemName] = mAcc
					}
					mAcc.count++
					mAcc.sumRating += float64(it.Rating)
					if it.Rating >= 4 {
						mAcc.positiveCount++
					} else {
						mAcc.negativeCount++
					}
					if it.Reason != nil && strings.TrimSpace(*it.Reason) != "" {
						trimmed := strings.TrimSpace(*it.Reason)
						if len(mAcc.sampleReasons) < 5 {
							mAcc.sampleReasons = append(mAcc.sampleReasons, trimmed)
						}
						processReasonTags(trimmed, tagCountsMap)
					}
				}
			}
		}
	}

	analytics := &dto.FeedbackAnalytics{
		TotalReviews:        totalReviews,
		RatingBreakdown:     make([]dto.RatingCount, 0, 5),
		BranchSummaries:     make([]dto.BranchRatingSummary, 0),
		TopMenuItems:        make([]dto.MenuItemRatingSummary, 0),
		NeedsAttentionItems: make([]dto.MenuItemRatingSummary, 0),
		CommonTags:          make([]dto.CommonTagCount, 0),
	}

	if totalReviews == 0 {
		for rVal := 5; rVal >= 1; rVal-- {
			analytics.RatingBreakdown = append(analytics.RatingBreakdown, dto.RatingCount{
				Rating:     rVal,
				Count:      0,
				Percentage: 0,
			})
		}
		return analytics, nil
	}

	analytics.AvgOverallRating = math.Round((sumOverall/float64(countOverall))*10) / 10
	if countResto > 0 {
		analytics.AvgRestoRating = math.Round((sumResto/float64(countResto))*10) / 10
	}
	if countApp > 0 {
		analytics.AvgAppRating = math.Round((sumApp/float64(countApp))*10) / 10
	}

	// Positive vs constructive
	posCount := ratingCounts[4] + ratingCounts[5]
	constructiveCount := ratingCounts[1] + ratingCounts[2] + ratingCounts[3]
	analytics.PositiveCount = posCount
	analytics.ConstructiveCount = constructiveCount
	analytics.SatisfactionRate = math.Round((float64(posCount)/float64(totalReviews))*1000) / 10

	// Breakdown 5 to 1
	for rVal := 5; rVal >= 1; rVal-- {
		c := ratingCounts[rVal]
		pct := math.Round((float64(c)/float64(totalReviews))*1000) / 10
		analytics.RatingBreakdown = append(analytics.RatingBreakdown, dto.RatingCount{
			Rating:     rVal,
			Count:      c,
			Percentage: pct,
		})
	}

	// Branch summaries
	for _, bAcc := range branchStatsMap {
		bSummary := dto.BranchRatingSummary{
			BranchID:     bAcc.id,
			BranchName:   bAcc.name,
			BranchSlug:   bAcc.slug,
			TotalReviews: bAcc.count,
		}
		if bAcc.count > 0 {
			bSummary.AvgOverallRating = math.Round((bAcc.sumOverall/float64(bAcc.count))*10) / 10
		}
		if bAcc.countResto > 0 {
			bSummary.AvgRestoRating = math.Round((bAcc.sumResto/float64(bAcc.countResto))*10) / 10
		}
		if bAcc.countApp > 0 {
			bSummary.AvgAppRating = math.Round((bAcc.sumApp/float64(bAcc.countApp))*10) / 10
		}
		analytics.BranchSummaries = append(analytics.BranchSummaries, bSummary)
	}
	sort.Slice(analytics.BranchSummaries, func(i, j int) bool {
		return analytics.BranchSummaries[i].TotalReviews > analytics.BranchSummaries[j].TotalReviews
	})

	// Menu items ranking
	menuList := make([]dto.MenuItemRatingSummary, 0, len(menuStatsMap))
	for _, mAcc := range menuStatsMap {
		if mAcc.count == 0 {
			continue
		}
		avg := math.Round((mAcc.sumRating/float64(mAcc.count))*10) / 10
		menuList = append(menuList, dto.MenuItemRatingSummary{
			ItemName:      mAcc.name,
			TotalReviews:  mAcc.count,
			AvgRating:     avg,
			PositiveCount: mAcc.positiveCount,
			NegativeCount: mAcc.negativeCount,
			SampleReasons: mAcc.sampleReasons,
		})
	}

	// Top rated items (>= 4.0, sorted by rating desc, then review count desc)
	sort.Slice(menuList, func(i, j int) bool {
		if menuList[i].AvgRating != menuList[j].AvgRating {
			return menuList[i].AvgRating > menuList[j].AvgRating
		}
		return menuList[i].TotalReviews > menuList[j].TotalReviews
	})
	topLimit := 5
	if len(menuList) < topLimit {
		topLimit = len(menuList)
	}
	analytics.TopMenuItems = append(analytics.TopMenuItems, menuList[:topLimit]...)

	// Needs attention items (items with negative reviews or lower rating, sorted by avg rating asc)
	needsAttention := make([]dto.MenuItemRatingSummary, 0)
	for _, it := range menuList {
		if it.NegativeCount > 0 || it.AvgRating < 4.2 {
			needsAttention = append(needsAttention, it)
		}
	}
	sort.Slice(needsAttention, func(i, j int) bool {
		if needsAttention[i].AvgRating != needsAttention[j].AvgRating {
			return needsAttention[i].AvgRating < needsAttention[j].AvgRating
		}
		return needsAttention[i].NegativeCount > needsAttention[j].NegativeCount
	})
	attentionLimit := 5
	if len(needsAttention) < attentionLimit {
		attentionLimit = len(needsAttention)
	}
	analytics.NeedsAttentionItems = append(analytics.NeedsAttentionItems, needsAttention[:attentionLimit]...)

	// Common tags
	tagList := make([]dto.CommonTagCount, 0, len(tagCountsMap))
	for _, tAcc := range tagCountsMap {
		tagList = append(tagList, dto.CommonTagCount{
			Tag:        tAcc.tag,
			Count:      tAcc.count,
			IsPositive: tAcc.isPositive,
		})
	}
	sort.Slice(tagList, func(i, j int) bool {
		return tagList[i].Count > tagList[j].Count
	})
	tagLimit := 12
	if len(tagList) < tagLimit {
		tagLimit = len(tagList)
	}
	analytics.CommonTags = append(analytics.CommonTags, tagList[:tagLimit]...)

	return analytics, nil
}

// GetRecentFeedbackForAI fetches recent reviews to provide rich context for Gemini analysis.
func (r *Repository) GetRecentFeedbackForAI(ctx context.Context, branchID *uuid.UUID, limit int) ([]model.OrderFeedbackAdminItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}
	items, _, err := r.ListOrderFeedbackAdmin(ctx, branchID, nil, "", limit, 0)
	return items, err
}

func processReasonTags(text string, tagMap map[string]*tagAccumulator) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	parts := strings.Split(text, ",")
	for _, p := range parts {
		cleaned := strings.TrimSpace(p)
		if len(cleaned) < 3 || len(cleaned) > 80 {
			continue
		}

		lower := strings.ToLower(cleaned)
		isPos := isPositiveTag(lower)

		tAcc, exists := tagMap[cleaned]
		if !exists {
			tAcc = &tagAccumulator{
				tag:        cleaned,
				isPositive: isPos,
			}
			tagMap[cleaned] = tAcc
		}
		tAcc.count++
	}
}

func isPositiveTag(t string) bool {
	posWords := []string{
		"lezat", "enak", "pas", "rapi", "hangat", "segar", "cepat", "ramah",
		"lancar", "puas", "bersih", "mantap", "empuk", "nagih", "otentik",
		"praktis", "modern", "bagus", "reorder",
	}
	for _, w := range posWords {
		if strings.Contains(t, w) {
			return true
		}
	}
	return false
}
