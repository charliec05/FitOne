package devseed

import (
	"context"
	"database/sql"
	"fmt"
)

// Seed populates the database with development fixtures.
func Seed(ctx context.Context, db *sql.DB) error {
	if err := seedUsers(ctx, db); err != nil {
		return err
	}
	if err := seedGyms(ctx, db); err != nil {
		return err
	}
	if err := seedMachines(ctx, db); err != nil {
		return err
	}
	if err := seedGymMachines(ctx, db); err != nil {
		return err
	}
	if err := seedGymPrices(ctx, db); err != nil {
		return err
	}
	if err := seedGymReviews(ctx, db); err != nil {
		return err
	}
	return nil
}

func seedUsers(ctx context.Context, db *sql.DB) error {
	const hashedPassword = "password"
	users := []struct {
		ID    string
		Email string
		Name  string
	}{
		{"11111111-1111-1111-1111-111111111111", "alex@example.com", "Alex Johnson"},
		{"22222222-2222-2222-2222-222222222222", "blake@example.com", "Blake Rivera"},
		{"33333333-3333-3333-3333-333333333333", "casey@example.com", "Casey Morgan"},
	}

	for _, user := range users {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO users (id, email, name, password)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO NOTHING
		`, user.ID, user.Email, user.Name, hashedPassword); err != nil {
			return fmt.Errorf("seed users: %w", err)
		}
	}
	return nil
}

func seedGyms(ctx context.Context, db *sql.DB) error {
	gyms := []struct {
		ID, Name, Address, Phone, Website string
		Lat, Lng                          float64
	}{
		{"44444444-4444-4444-4444-444444444441", "Pike Place Fitness", "1912 Pike Pl, Seattle, WA", "+1-206-555-0101", "https://pikeplacefitness.example", 47.6097, -122.3425},
		{"44444444-4444-4444-4444-444444444442", "Capitol Hill Strength", "1423 10th Ave, Seattle, WA", "+1-206-555-0102", "https://capitolhillstrength.example", 47.6133, -122.3185},
		{"44444444-4444-4444-4444-444444444443", "Ballard Ironworks", "5317 Ballard Ave NW, Seattle, WA", "+1-206-555-0103", "https://ballardironworks.example", 47.6687, -122.3845},
		{"44444444-4444-4444-4444-444444444444", "Fremont Powerhouse", "3601 Fremont Ave N, Seattle, WA", "+1-206-555-0104", "https://fremontpowerhouse.example", 47.6529, -122.3505},
		{"44444444-4444-4444-4444-444444444445", "South Lake Union Athletics", "420 Pontius Ave N, Seattle, WA", "+1-206-555-0105", "https://sluathletics.example", 47.6230, -122.3335},
	}

	for _, gym := range gyms {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO gyms (id, name, lat, lng, address, phone, website)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING
		`, gym.ID, gym.Name, gym.Lat, gym.Lng, gym.Address, gym.Phone, gym.Website); err != nil {
			return fmt.Errorf("seed gyms: %w", err)
		}
	}

	return nil
}

func seedMachines(ctx context.Context, db *sql.DB) error {
	machines := []struct {
		ID, Name, BodyPart string
	}{
		{"55555555-0000-0000-0000-000000000001", "Leg Press", "Legs"},
		{"55555555-0000-0000-0000-000000000002", "Smith Machine", "Full Body"},
		{"55555555-0000-0000-0000-000000000003", "Lat Pulldown", "Back"},
		{"55555555-0000-0000-0000-000000000004", "Cable Row", "Back"},
		{"55555555-0000-0000-0000-000000000005", "Treadmill", "Cardio"},
		{"55555555-0000-0000-0000-000000000006", "Elliptical", "Cardio"},
		{"55555555-0000-0000-0000-000000000007", "Bench Press", "Chest"},
		{"55555555-0000-0000-0000-000000000008", "Squat Rack", "Legs"},
		{"55555555-0000-0000-0000-000000000009", "Rowing Machine", "Cardio"},
		{"55555555-0000-0000-0000-000000000010", "Dumbbells", "Arms"},
	}

	for _, machine := range machines {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO machines (id, name, body_part)
			VALUES ($1, $2, $3)
			ON CONFLICT (id) DO NOTHING
		`, machine.ID, machine.Name, machine.BodyPart); err != nil {
			return fmt.Errorf("seed machines: %w", err)
		}
	}

	return nil
}

func seedGymMachines(ctx context.Context, db *sql.DB) error {
	pairs := []struct {
		GymID, MachineID string
		Quantity        int
	}{
		{"44444444-4444-4444-4444-444444444441", "55555555-0000-0000-0000-000000000001", 4},
		{"44444444-4444-4444-4444-444444444441", "55555555-0000-0000-0000-000000000007", 3},
		{"44444444-4444-4444-4444-444444444441", "55555555-0000-0000-0000-000000000010", 20},
		{"44444444-4444-4444-4444-444444444442", "55555555-0000-0000-0000-000000000002", 3},
		{"44444444-4444-4444-4444-444444444442", "55555555-0000-0000-0000-000000000003", 2},
		{"44444444-4444-4444-4444-444444444443", "55555555-0000-0000-0000-000000000005", 6},
		{"44444444-4444-4444-4444-444444444443", "55555555-0000-0000-0000-000000000004", 3},
		{"44444444-4444-4444-4444-444444444444", "55555555-0000-0000-0000-000000000008", 2},
		{"44444444-4444-4444-4444-444444444444", "55555555-0000-0000-0000-000000000009", 2},
		{"44444444-4444-4444-4444-444444444445", "55555555-0000-0000-0000-000000000006", 4},
		{"44444444-4444-4444-4444-444444444445", "55555555-0000-0000-0000-000000000001", 3},
	}

	for _, pair := range pairs {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO gym_machines (gym_id, machine_id, quantity)
			VALUES ($1, $2, $3)
			ON CONFLICT (gym_id, machine_id) DO NOTHING
		`, pair.GymID, pair.MachineID, pair.Quantity); err != nil {
			return fmt.Errorf("seed gym machines: %w", err)
		}
	}

	return nil
}

func seedGymPrices(ctx context.Context, db *sql.DB) error {
	plans := []struct {
		GymID, Plan, Period string
		Price              int
	}{
		{"44444444-4444-4444-4444-444444444441", "Monthly", "month", 8900},
		{"44444444-4444-4444-4444-444444444441", "Quarterly", "quarter", 24900},
		{"44444444-4444-4444-4444-444444444441", "Annual", "year", 89900},
		{"44444444-4444-4444-4444-444444444442", "Monthly", "month", 9900},
		{"44444444-4444-4444-4444-444444444442", "Annual", "year", 94900},
		{"44444444-4444-4444-4444-444444444443", "Monthly", "month", 7500},
		{"44444444-4444-4444-4444-444444444443", "Annual", "year", 79900},
		{"44444444-4444-4444-4444-444444444444", "Monthly", "month", 9200},
		{"44444444-4444-4444-4444-444444444445", "Monthly", "month", 10500},
		{"44444444-4444-4444-4444-444444444445", "Corporate", "month", 9500},
	}

	for _, plan := range plans {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO gym_prices (gym_id, plan_name, price_cents, period)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (gym_id, plan_name) DO NOTHING
		`, plan.GymID, plan.Plan, plan.Price, plan.Period); err != nil {
			return fmt.Errorf("seed gym prices: %w", err)
		}
	}

	return nil
}

func seedGymReviews(ctx context.Context, db *sql.DB) error {
	reviews := []struct {
		ID, GymID, UserID, Comment string
		Rating                      int
	}{
		{"77777777-0000-0000-0000-000000000001", "44444444-4444-4444-4444-444444444441", "11111111-1111-1111-1111-111111111111", "Spacious facility with great cardio equipment.", 5},
		{"77777777-0000-0000-0000-000000000002", "44444444-4444-4444-4444-444444444441", "22222222-2222-2222-2222-222222222222", "Friendly staff and clean locker rooms.", 4},
		{"77777777-0000-0000-0000-000000000003", "44444444-4444-4444-4444-444444444442", "11111111-1111-1111-1111-111111111111", "Peak hours can be busy but equipment variety is excellent.", 4},
		{"77777777-0000-0000-0000-000000000004", "44444444-4444-4444-4444-444444444442", "33333333-3333-3333-3333-333333333333", "Love the strength training classes here!", 5},
		{"77777777-0000-0000-0000-000000000005", "44444444-4444-4444-4444-444444444443", "22222222-2222-2222-2222-222222222222", "Plenty of parking and open 24 hours.", 4},
		{"77777777-0000-0000-0000-000000000006", "44444444-4444-4444-4444-444444444443", "33333333-3333-3333-3333-333333333333", "Could use more squat racks but still solid.", 4},
		{"77777777-0000-0000-0000-000000000007", "44444444-4444-4444-4444-444444444444", "11111111-1111-1111-1111-111111111111", "Great vibe and group classes.", 5},
		{"77777777-0000-0000-0000-000000000008", "44444444-4444-4444-4444-444444444444", "22222222-2222-2222-2222-222222222222", "Wish they had more towel service.", 3},
		{"77777777-0000-0000-0000-000000000009", "44444444-4444-4444-4444-444444444445", "33333333-3333-3333-3333-333333333333", "Brand new equipment and helpful trainers.", 5},
		{"77777777-0000-0000-0000-000000000010", "44444444-4444-4444-4444-444444444445", "11111111-1111-1111-1111-111111111111", "Locker rooms could be larger but overall great.", 4},
	}

	for _, review := range reviews {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO gym_reviews (id, gym_id, user_id, rating, comment)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING
		`, review.ID, review.GymID, review.UserID, review.Rating, sql.NullString{String: review.Comment, Valid: review.Comment != ""}); err != nil {
			return fmt.Errorf("seed gym reviews: %w", err)
		}
	}

	return nil
}
