package main

import (
	"flag"
	"fmt"
	"groups-parser/pkg/database"
	"groups-parser/pkg/database/models"
	"groups-parser/pkg/dto"
	"groups-parser/pkg/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	DSN := flag.String("dsn", "", "database connection string")
	flag.Parse()

	if *DSN == "" {
		panic("dsn is required")
	}

	err := database.InitDatabase(*DSN)
	if err != nil {
		panic(err)
	}

	data, err := http.FetchDataFromURL[dto.GroupResponseDTO](
		"https://dec.mgutm.ru/api/groups",
		map[string]string{
			"year": "2025-2026",
		},
	)
	if err != nil {
		panic(err)
	}

	levelsMap := make(map[int]string)
	for _, level := range data.Data.LevelEducation {
		levelsMap[level.Id] = level.Name
	}

	numWorkers := 20
	groupCh := make(chan dto.GroupDTO, len(data.Data.Groups))
	resultCh := make(chan models.Group, len(data.Data.Groups))
	var wg sync.WaitGroup

	for _, group := range data.Data.Groups {
		groupCh <- group
	}
	close(groupCh)

	worker := func() {
		defer wg.Done()
		for group := range groupCh {
			groupInfo, err := http.FetchDataFromURL[dto.GroupInfoDTO](
				"https://dec.mgutm.ru/api/UserInfo/GroupInfo",
				map[string]string{
					"groupID": strconv.Itoa(group.GroupID),
				},
			)
			if err != nil {
				panic(err)
			}

			parsedGroup := models.Group{
				AddedDate: time.Now().UTC(),
				ID:        uint(group.GroupID),
				Years:     groupInfo.Data.GroupYear,
				Name:      group.GroupName,
				Specialty: groupInfo.Data.SpecialName,
				Level:     groupInfo.Data.LevelName,
				Course:    group.Course,
				Abbreviation: func() string {
					facultyName := strings.ReplaceAll(groupInfo.Data.FaculName, "(филиал)", "")
					words := strings.Fields(facultyName)
					abbr := ""
					for _, word := range words {
						if len(word) == 0 {
							continue
						}
						runes := []rune(word)
						if len(runes) == 1 {
							abbr += string(runes[0])
						} else {
							abbr += strings.ToUpper(string(runes[0]))
						}
					}
					return abbr
				}(),
				Faculty: groupInfo.Data.FaculName,
			}

			resultCh <- parsedGroup
		}
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	wg.Wait()
	close(resultCh)

	var groups []models.Group
	for g := range resultCh {
		groups = append(groups, g)
	}

	for _, g := range groups {
		var existing models.Group
		result := database.GetDB().Where("id = ?", g.ID).First(&existing)
		if result.RowsAffected > 0 {
			fmt.Println("Duplicate group found:", g.ID, g.Name)
			continue
		}

		if err := database.GetDB().Create(&g).Error; err != nil {
			fmt.Println("Error saving group:", g.ID, err.Error())
		}
	}

	database.CloseDatabase()
}
