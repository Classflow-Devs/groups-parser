package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"
	"unicode"

	"groups-parser/pkg/database"
	"groups-parser/pkg/database/models"
	"groups-parser/pkg/dto"

	shttp "github.com/krispeckt/simple-fasthttp"
)

const (
	defaultBranchID = uint(1)
	maxRetries      = 3
	retryDelay      = 2 * time.Second
)

type parsedGroup struct {
	id        uint
	name      string
	course    uint8
	faculty   string
	specialty string
	level     string
}

func main() {
	dsn := flag.String("dsn", "", "PostgreSQL connection string")
	year := flag.String("year", "2025-2026", "academic year (e.g. 2025-2026)")
	numWorkers := flag.Int("workers", 20, "number of concurrent HTTP workers")
	migrate := flag.Bool("migrate", false, "run AutoMigrate before parsing")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("--dsn is required")
	}

	if err := database.InitDatabase(*dsn); err != nil {
		log.Fatal("init db:", err)
	}
	defer database.CloseDatabase()

	db := database.GetDB()

	if *migrate {
		err := db.AutoMigrate(
			&models.Branch{},
			&models.Faculty{},
			&models.Speciality{},
			&models.Group{},
		)
		if err != nil {
			log.Fatal("migrate:", err)
		}
		log.Println("migration completed")
	}

	// Load all branches for auto-detection by country code
	var branches []models.Branch
	if err := db.Find(&branches).Error; err != nil {
		log.Fatal("load branches:", err)
	}
	branchByCity := make(map[string]uint, len(branches))
	for _, b := range branches {
		branchByCity[b.City] = b.ID
	}
	log.Printf("loaded %d branches", len(branches))

	// Resolve branch by first 2 letters of group name → branch.Country.
	// Falls back to defaultBranchID when no match.
	resolveBranch := func(groupName string) uint {
		if prefix := extractPrefix(groupName); prefix != "" {
			if id, ok := branchByCity[prefix]; ok {
				return id
			}
		}
		return defaultBranchID
	}

	// Fetch groups list (with retry)
	resp, err := retryGet[dto.GroupResponseDTO](
		"https://dec.mgutm.ru/api/groups",
		map[string]string{"year": *year},
	)
	if err != nil {
		log.Fatal("fetch groups:", err)
	}
	log.Printf("fetched %d groups for year %s", len(resp.Data.Groups), *year)

	// Fan-out: fetch group infos concurrently
	groupCh := make(chan dto.GroupDTO, len(resp.Data.Groups))
	resultCh := make(chan parsedGroup, len(resp.Data.Groups))

	for _, g := range resp.Data.Groups {
		groupCh <- g
	}
	close(groupCh)

	var wg sync.WaitGroup
	wg.Add(*numWorkers)
	for i := 0; i < *numWorkers; i++ {
		go func() {
			defer wg.Done()
			for g := range groupCh {
				info, err := retryGet[dto.GroupInfoDTO](
					"https://dec.mgutm.ru/api/UserInfo/GroupInfo",
					map[string]string{"groupID": strconv.Itoa(g.GroupID)},
				)
				if err != nil {
					log.Printf("WARN: skip group %d %q — %v", g.GroupID, g.GroupName, err)
					continue
				}
				resultCh <- parsedGroup{
					id:        uint(g.GroupID),
					name:      g.GroupName,
					course:    uint8(g.Course),
					faculty:   info.Data.FaculName,
					specialty: info.Data.SpecialName,
					level:     info.Data.LevelName,
				}

				time.Sleep(time.Millisecond * 100)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var groups []parsedGroup
	for pg := range resultCh {
		groups = append(groups, pg)
	}
	log.Printf("fetched detailed info for %d/%d groups", len(groups), len(resp.Data.Groups))

	// Process serially: upsert faculties → specialities → groups
	type cacheKey struct{ a, b string }
	facultyCache := make(map[cacheKey]uint)
	specialityCache := make(map[cacheKey]uint)

	saved, skipped, failed := 0, 0, 0

	for _, pg := range groups {
		branchID := resolveBranch(pg.name)

		// Upsert faculty
		fKey := cacheKey{pg.faculty, strconv.Itoa(int(branchID))}
		facultyID, ok := facultyCache[fKey]
		if !ok {
			var faculty models.Faculty
			err := db.Where("name = ? AND branch_id = ?", pg.faculty, branchID).First(&faculty).Error
			if err != nil {
				faculty = models.Faculty{Name: pg.faculty, BranchID: branchID}
				if err := db.Create(&faculty).Error; err != nil {
					log.Printf("ERROR: create faculty %q: %v", pg.faculty, err)
					failed++
					continue
				}
			}
			facultyID = faculty.ID
			facultyCache[fKey] = facultyID
		}

		// Upsert speciality — explicit WHERE to avoid GORM ignoring empty-string fields
		spKey := cacheKey{pg.specialty + "|" + pg.level, strconv.Itoa(int(branchID))}
		specialityID, ok := specialityCache[spKey]
		if !ok {
			var speciality models.Speciality
			err := db.Where("name = ? AND level = ? AND branch_id = ?", pg.specialty, pg.level, branchID).
				First(&speciality).Error
			if err != nil {
				speciality = models.Speciality{Name: pg.specialty, Level: pg.level, BranchID: branchID}
				if err := db.Create(&speciality).Error; err != nil {
					log.Printf("ERROR: create speciality %q: %v", pg.specialty, err)
					failed++
					continue
				}
			}
			specialityID = speciality.ID
			specialityCache[spKey] = specialityID
		}

		// Skip already existing groups
		var existing models.Group
		if db.Select("id").First(&existing, pg.id).RowsAffected > 0 {
			skipped++
			continue
		}

		group := models.Group{
			ID:           pg.id,
			BranchID:     branchID,
			Name:         pg.name,
			Course:       pg.course,
			FacultyID:    facultyID,
			SpecialityID: specialityID,
		}
		if err := db.Create(&group).Error; err != nil {
			log.Printf("ERROR: create group %d %q: %v", pg.id, pg.name, err)
			failed++
			continue
		}
		saved++
	}

	log.Printf("done — saved: %d  skipped: %d  errors: %d", saved, skipped, failed)
}

// extractPrefix returns the first 2 uppercase letters from s.
// Returns "" if fewer than 2 letters are found.
func extractPrefix(s string) string {
	var letters []rune
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters = append(letters, unicode.ToUpper(r))
			if len(letters) == 2 {
				break
			}
		}
	}
	if len(letters) < 2 {
		return ""
	}
	return string(letters)
}

// retryGet performs GET with exponential back-off retries.
func retryGet[T any](rawURL string, params map[string]string) (*T, error) {
	u, _ := url.Parse(rawURL)
	for attempt := 1; attempt <= maxRetries; attempt++ {
		urlParams := url.Values{}
		for k, v := range params {
			urlParams[k] = []string{v}
		}
		u.RawQuery = urlParams.Encode()

		result, _, err := shttp.Get[T](context.Background(), u, nil, nil)
		if err == nil {
			return result, nil
		}
		log.Printf("WARN: attempt %d/%d GET %s: %v", attempt, maxRetries, u.RequestURI(), err)
		if attempt < maxRetries {
			time.Sleep(retryDelay * time.Duration(attempt))
		}
	}
	return nil, fmt.Errorf("all %d attempts failed for %s", maxRetries, rawURL)
}
