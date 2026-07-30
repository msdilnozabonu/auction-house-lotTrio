package lots

const (
	lotExpired = `SELECT id FROM lots WHERE status = 'live' AND ends_at < $1`
	closeLot   = `UPDATE lots 
	SET status = 'closed',
	    current_winner_id = (SELECT bidder_id FROM bids WHERE lot_id = $1
		ORDER BY amount DESC, created_at DESC LIMIT 1)
	WHERE id = $1 AND status = 'live'`
	selectByID = `SELECT id, title, description, category, start_price, current_price, COALESCE(current_winner_id, 0), 
       status, starts_at, ends_at, photo_path, seller_id FROM lots WHERE id = $1`
	selectByIDForBid = `SELECT id, title, description, category, start_price, current_price, 
       COALESCE(current_winner_id, 0), status, starts_at, ends_at, photo_path, seller_id FROM lots WHERE id = $1 
       AND status = 'live'`
	selectAll = `SELECT id, seller_id, title, coalesce(description,''), category, start_price, current_price, 
       COALESCE(current_winner_id, 0), status, starts_at, ends_at, coalesce(photo_path, '') 
		FROM lots WHERE status = 'live' 
      	AND ($1 = '' OR title ILIKE '%'||$1||'%' OR description ILIKE '%'||$1||'%')
      	AND ($2 = '' OR category = $2) AND ($3 = 0 OR current_price >= $3)
      	AND ($4 = 0 OR current_price <= $4) ORDER BY id LIMIT $5 OFFSET $6`
	countAll = `SELECT COUNT(*) FROM lots WHERE status = 'live' 
        AND ($1 = '' OR title ILIKE '%'||$1||'%' OR description ILIKE '%'||$1||'%')
        AND ($2 = '' OR category = $2) AND ($3 = 0 OR current_price >= $3) AND ($4 = 0 OR current_price <= $4)`
	baseQuery = `SELECT id, seller_id, title, coalesce(description, ''), start_price, current_price, 
       coalesce(current_winner_id, 0), status, starts_at, ends_at, coalesce(photo_path, '') FROM lots WHERE 1=1`
	baseCountQuery = `SELECT COUNT(*) FROM lots WHERE 1=1`
	moderateLot    = `UPDATE lots SET moderation_status = $1, rejection_reason = $2 
            WHERE id = $3 AND moderation_status = 'pending'`
	lotsByStatus = `SELECT status, COUNT(*) FROM lots GROUP BY status`
	revenue      = `SELECT COALESCE(SUM(current_price), 0), COALESCE(AVG(current_price), 0)
		FROM lots WHERE status = 'closed' AND current_winner_id IS NOT NULL`
	topCategories = `SELECT category, COUNT(*), COALESCE(SUM(current_price), 0) FROM lots WHERE status = 'closed'
		AND current_winner_id IS NOT NULL GROUP BY category ORDER BY SUM(current_price) DESC LIMIT 5`
	lotExport = `SELECT id, title, category, status, seller_id, current_price, ends_at 
	FROM lots WHERE %s BETWEEN $1 AND $2 ORDER BY id`
	selectMine = `Select id, seller_id, title, description,category, start_price, current_price,
       COALESCE(current_winner_id, 0), status, starts_at, ends_at, photo_path FROM lots WHERE seller_id = $1 
      	AND ($2 = '' OR title ILIKE '%'||$2||'%' OR description ILIKE '%'||$2||'%')
      	AND ($3 = '' OR category = $3) AND ($4 = 0 OR current_price >= $4)
      	AND ($5 = 0 OR current_price <= $5) ORDER BY id LIMIT $6 OFFSET $7`
	cancelLot = `UPDATE lots SET status = 'cancelled', cancellation_reason = $1
            WHERE id = $2 AND status IN('draft', 'live')`
	createReport = `INSERT INTO reports (lot_id, reporter_id, reason)
				VALUES ($1, $2, $3)`
	getListOfReports = `SELECT id, lot_id, reporter_id, reason, created_at 
		FROM reports WHERE status = 'pending' ORDER BY created_at`
)
