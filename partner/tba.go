// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game
//
// Methods for publishing data to and retrieving data from The Blue Alliance.

package partner

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/mitchellh/mapstructure"
)

const (
	tbaBaseUrl = "https://www.thebluealliance.com"
	tbaAuthKey = "MAApv9MCuKY9MSFkXLuzTSYBCdosboxDq8Q3ujUE2Mn8PD3Nmv64uczu5Lvy0NQ3"
	AvatarsDir = "static/img/avatars"
)

type TbaClient struct {
	BaseUrl         string
	eventCode       string
	secretId        string
	secret          string
	eventNamesCache map[string]string
}

type TbaMatch struct {
	CompLevel      string                    `json:"comp_level"`
	SetNumber      int                       `json:"set_number"`
	MatchNumber    int                       `json:"match_number"`
	Alliances      map[string]*TbaAlliance   `json:"alliances"`
	ScoreBreakdown map[string]map[string]any `json:"score_breakdown"`
	TimeString     string                    `json:"time_string"`
	TimeUtc        string                    `json:"time_utc"`
	DisplayName    string                    `json:"display_name"`
}

type TbaAlliance struct {
	Teams      []string `json:"teams"`
	Surrogates []string `json:"surrogates"`
	Dqs        []string `json:"dqs"`
	Score      *int     `json:"score"`
}

type TbaHubScore struct {
	AutoCount        int `mapstructure:"autoCount"`
	AutoPoints       int `mapstructure:"autoPoints"`
	EndgameCount     int `mapstructure:"endgameCount"`
	EndgamePoints    int `mapstructure:"endgamePoints"`
	Shift1Count      int `mapstructure:"shift1Count"`
	Shift1Points     int `mapstructure:"shift1Points"`
	Shift2Count      int `mapstructure:"shift2Count"`
	Shift2Points     int `mapstructure:"shift2Points"`
	Shift3Count      int `mapstructure:"shift3Count"`
	Shift3Points     int `mapstructure:"shift3Points"`
	Shift4Count      int `mapstructure:"shift4Count"`
	Shift4Points     int `mapstructure:"shift4Points"`
	TeleopCount      int `mapstructure:"teleopCount"`
	TeleopPoints     int `mapstructure:"teleopPoints"`
	TotalCount       int `mapstructure:"totalCount"`
	TotalPoints      int `mapstructure:"totalPoints"`
	TransitionCount  int `mapstructure:"transitionCount"`
	TransitionPoints int `mapstructure:"transitionPoints"`
}

// 2026 REBUILT Score Breakdown Structure
type TbaScoreBreakdown struct {
	AdjustPoints int `mapstructure:"adjustPoints"`

	// Auto
	AutoTowerRobot1 string `mapstructure:"autoTowerRobot1"`
	AutoTowerRobot2 string `mapstructure:"autoTowerRobot2"`
	AutoTowerRobot3 string `mapstructure:"autoTowerRobot3"`
	AutoTowerPoints int    `mapstructure:"autoTowerPoints"`
	TotalAutoPoints int    `mapstructure:"totalAutoPoints"`

	// Teleop
	TotalTeleopPoints int `mapstructure:"totalTeleopPoints"`

	// Endgame
	EndGameTowerRobot1 string `mapstructure:"endGameTowerRobot1"`
	EndGameTowerRobot2 string `mapstructure:"endGameTowerRobot2"`
	EndGameTowerRobot3 string `mapstructure:"endGameTowerRobot3"`
	EndGameTowerPoints int    `mapstructure:"endGameTowerPoints"`

	// Totals
	TotalTowerPoints int `mapstructure:"totalTowerPoints"`
	TotalPoints      int `mapstructure:"totalPoints"`

	// Fouls
	MinorFoulCount int  `mapstructure:"minorFoulCount"`
	MajorFoulCount int  `mapstructure:"majorFoulCount"`
	G206Penalty    bool `mapstructure:"g206Penalty"` // RP Collusion
	FoulPoints     int  `mapstructure:"foulPoints"`

	// Ranking Points & Achievements
	EnergizedAchieved    bool `mapstructure:"energizedAchieved"`
	SuperchargedAchieved bool `mapstructure:"superchargedAchieved"`
	TraversalAchieved    bool `mapstructure:"traversalAchieved"`
	RP                   int  `mapstructure:"rp"`

	// 2026 Hub Score Object
	HubScore TbaHubScore `mapstructure:"hubScore"`
}

type TbaRanking struct {
	TeamKey string  `json:"team_key"`
	Rank    int     `json:"rank"`
	RP      float32 `json:"RP"`
	Match   float32 `json:"Match"` // Avg Match Points
	Auto    float32 `json:"Auto"`  // Avg Auto Points
	Tower   float32 `json:"Tower"` // Avg Tower Points (Replaces Barge)
	Wins    int     `json:"wins"`
	Losses  int     `json:"losses"`
	Ties    int     `json:"ties"`
	Dqs     int     `json:"dqs"`
	Played  int     `json:"played"`
}

type TbaRankings struct {
	Breakdowns []string     `json:"breakdowns"`
	Rankings   []TbaRanking `json:"rankings"`
}

type TbaTeam struct {
	TeamNumber int    `json:"team_number"`
	Name       string `json:"name"`
	Nickname   string `json:"nickname"`
	City       string `json:"city"`
	StateProv  string `json:"state_prov"`
	Country    string `json:"country"`
	RookieYear int    `json:"rookie_year"`
}

type TbaRobot struct {
	RobotName string `json:"robot_name"`
	Year      int    `json:"year"`
}

type TbaAward struct {
	Name      string `json:"name"`
	EventKey  string `json:"event_key"`
	Year      int    `json:"year"`
	EventName string
}

type TbaEvent struct {
	Name string `json:"name"`
}

type TbaMediaItem struct {
	Details map[string]any `json:"details"`
	Type    string         `json:"type"`
}

type TbaPublishedAward struct {
	Name    string `json:"name_str"`
	TeamKey string `json:"team_key"`
	Awardee string `json:"awardee"`
}

// 2026 Mappings
var autoTowerMapping = map[bool]string{false: "None", true: "Level1"}
var endGameStatusMapping = map[game.EndgameStatus]string{
	game.EndgameNone:   "None",
	game.EndgameLevel1: "Level1",
	game.EndgameLevel2: "Level2",
	game.EndgameLevel3: "Level3",
}

func NewTbaClient(eventCode, secretId, secret string) *TbaClient {
	return &TbaClient{
		BaseUrl:         tbaBaseUrl,
		eventCode:       eventCode,
		secretId:        secretId,
		secret:          secret,
		eventNamesCache: make(map[string]string),
	}
}

func (client *TbaClient) GetTeam(teamNumber int) (*TbaTeam, error) {
	path := fmt.Sprintf("/api/v3/team/%s", getTbaTeam(teamNumber))
	resp, err := client.getRequest(path)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var teamData TbaTeam
	err = json.Unmarshal(body, &teamData)

	return &teamData, err
}

func (client *TbaClient) GetRobotName(teamNumber int, year int) (string, error) {
	path := fmt.Sprintf("/api/v3/team/%s/robots", getTbaTeam(teamNumber))
	resp, err := client.getRequest(path)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var robots []*TbaRobot
	err = json.Unmarshal(body, &robots)
	if err != nil {
		return "", err
	}
	for _, robot := range robots {
		if robot.Year == year {
			return robot.RobotName, nil
		}
	}
	return "", nil
}

func (client *TbaClient) GetTeamAwards(teamNumber int) ([]*TbaAward, error) {
	path := fmt.Sprintf("/api/v3/team/%s/awards", getTbaTeam(teamNumber))
	resp, err := client.getRequest(path)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var awards []*TbaAward
	err = json.Unmarshal(body, &awards)
	if err != nil {
		return nil, err
	}

	for _, award := range awards {
		if _, ok := client.eventNamesCache[award.EventKey]; !ok {
			client.eventNamesCache[award.EventKey], err = client.getEventName(award.EventKey)
			if err != nil {
				return nil, err
			}
		}
		award.EventName = client.eventNamesCache[award.EventKey]
	}

	return awards, nil
}

func (client *TbaClient) DownloadTeamAvatar(teamNumber, year int) error {
	path := fmt.Sprintf("/api/v3/team/%s/media/%d", getTbaTeam(teamNumber), year)
	resp, err := client.getRequest(path)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var mediaItems []*TbaMediaItem
	err = json.Unmarshal(body, &mediaItems)
	if err != nil {
		return err
	}

	for _, item := range mediaItems {
		if item.Type == "avatar" {
			base64String, ok := item.Details["base64Image"].(string)
			if !ok {
				return fmt.Errorf("Could not interpret avatar response from TBA: %v", item)
			}
			avatarBytes, err := base64.StdEncoding.DecodeString(base64String)
			if err != nil {
				return err
			}

			avatarPath := fmt.Sprintf("%s/%d.png", AvatarsDir, teamNumber)
			return os.WriteFile(avatarPath, avatarBytes, 0644)
		}
	}

	return nil
}

func (client *TbaClient) PublishTeams(database *model.Database) error {
	teams, err := database.GetAllTeams()
	if err != nil {
		return err
	}

	teamKeys := make([]string, len(teams))
	for i, team := range teams {
		teamKeys[i] = getTbaTeam(team.Id)
	}
	jsonBody, err := json.Marshal(teamKeys)
	if err != nil {
		return err
	}

	resp, err := client.postRequest("team_list", "update", jsonBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

func (client *TbaClient) PublishMatches(database *model.Database) error {
	qualMatches, err := database.GetMatchesByType(model.Qualification, false)
	if err != nil {
		return err
	}
	playoffMatches, err := database.GetMatchesByType(model.Playoff, false)
	if err != nil {
		return err
	}
	eventSettings, err := database.GetEventSettings()
	if err != nil {
		return err
	}
	matches := append(qualMatches, playoffMatches...)
	tbaMatches := make([]TbaMatch, len(matches))

	for i, match := range matches {
		var scoreBreakdown map[string]map[string]any
		var redScore, blueScore *int
		var redCards, blueCards map[string]string
		if match.IsComplete() {
			matchResult, err := database.GetMatchResultForMatch(match.Id)
			if err != nil {
				return err
			}
			if matchResult != nil {
				scoreBreakdown = make(map[string]map[string]any)
				scoreBreakdown["red"] = createTbaScoringBreakdown(eventSettings, &match, matchResult, "red")
				scoreBreakdown["blue"] = createTbaScoringBreakdown(eventSettings, &match, matchResult, "blue")
				redScoreValue := scoreBreakdown["red"]["totalPoints"].(int)
				blueScoreValue, _ := scoreBreakdown["blue"]["totalPoints"].(int)
				redScore = &redScoreValue
				blueScore = &blueScoreValue
				redCards = matchResult.RedCards
				blueCards = matchResult.BlueCards
			}
		}
		alliances := make(map[string]*TbaAlliance)
		alliances["red"] = createTbaAlliance(
			[3]int{match.Red1, match.Red2, match.Red3},
			[3]bool{match.Red1IsSurrogate, match.Red2IsSurrogate, match.Red3IsSurrogate},
			redScore,
			redCards,
		)
		alliances["blue"] = createTbaAlliance(
			[3]int{match.Blue1, match.Blue2, match.Blue3},
			[3]bool{match.Blue1IsSurrogate, match.Blue2IsSurrogate, match.Blue3IsSurrogate},
			blueScore,
			blueCards,
		)

		tbaMatches[i] = TbaMatch{
			CompLevel:      match.TbaMatchKey.CompLevel,
			SetNumber:      match.TbaMatchKey.SetNumber,
			MatchNumber:    match.TbaMatchKey.MatchNumber,
			Alliances:      alliances,
			ScoreBreakdown: scoreBreakdown,
			TimeString:     match.Time.Local().Format("3:04 PM"),
			TimeUtc:        match.Time.UTC().Format("2006-01-02T15:04:05"),
		}
	}
	jsonBody, err := json.Marshal(tbaMatches)
	if err != nil {
		return err
	}

	resp, err := client.postRequest("matches", "update", jsonBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

func (client *TbaClient) PublishRankings(database *model.Database) error {
	rankings, err := database.GetAllRankings()
	if err != nil {
		return err
	}

	// 2026 Ranking Headers: RP, Match, Auto, Tower
	breakdowns := []string{"RP", "Match", "Auto", "Tower"}
	tbaRankings := make([]TbaRanking, len(rankings))
	for i, ranking := range rankings {
		tbaRankings[i] = TbaRanking{
			TeamKey: getTbaTeam(ranking.TeamId),
			Rank:    ranking.Rank,
			RP:      float32(ranking.RankingPoints) / float32(ranking.Played),
			Match:   float32(ranking.MatchPoints) / float32(ranking.Played),
			Auto:    float32(ranking.AutoPoints) / float32(ranking.Played),
			// 修正點 1: 使用 TowerPoints 取代 TotalTowerPoints
			Tower:  float32(ranking.TowerPoints) / float32(ranking.Played),
			Wins:   ranking.Wins,
			Losses: ranking.Losses,
			Ties:   ranking.Ties,
			Dqs:    ranking.Disqualifications,
			Played: ranking.Played,
		}
	}
	jsonBody, err := json.Marshal(TbaRankings{breakdowns, tbaRankings})
	if err != nil {
		return err
	}

	resp, err := client.postRequest("rankings", "update", jsonBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

func (client *TbaClient) PublishAlliances(database *model.Database) error {
	alliances, err := database.GetAllAlliances()
	if err != nil {
		return err
	}

	tbaAlliances := make([][]string, len(alliances))
	for i, alliance := range alliances {
		for _, allianceTeamId := range alliance.TeamIds {
			tbaAlliances[i] = append(tbaAlliances[i], getTbaTeam(allianceTeamId))
		}
	}
	jsonBody, err := json.Marshal(tbaAlliances)
	if err != nil {
		return err
	}

	resp, err := client.postRequest("alliance_selections", "update", jsonBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}

	eventSettings, err := database.GetEventSettings()
	if err != nil {
		return err
	}
	playoffType := 0
	if eventSettings.PlayoffType == model.DoubleEliminationPlayoff {
		playoffType = 10
	}
	resp, err = client.postRequest("info", "update", []byte(fmt.Sprintf("{\"playoff_type\":%d}", playoffType)))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}

	return nil
}

func (client *TbaClient) PublishAwards(database *model.Database) error {
	awards, err := database.GetAllAwards()
	if err != nil {
		return err
	}

	tbaAwards := make([]TbaPublishedAward, len(awards))
	for i, award := range awards {
		tbaAwards[i].Name = award.AwardName
		tbaAwards[i].TeamKey = getTbaTeam(award.TeamId)
		tbaAwards[i].Awardee = award.PersonName
	}
	jsonBody, err := json.Marshal(tbaAwards)
	if err != nil {
		return err
	}

	resp, err := client.postRequest("awards", "update", jsonBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

func (client *TbaClient) DeletePublishedMatches() error {
	resp, err := client.postRequest("matches", "delete_all", []byte(client.eventCode))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

func (client *TbaClient) getEventName(eventCode string) (string, error) {
	path := fmt.Sprintf("/api/v3/event/%s", eventCode)
	resp, err := client.getRequest(path)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var event TbaEvent
	err = json.Unmarshal(body, &event)
	if err != nil {
		return "", err
	}

	return event.Name, err
}

func getTbaTeam(team int) string {
	return fmt.Sprintf("frc%d", team)
}

func (client *TbaClient) getRequest(path string) (*http.Response, error) {
	url := client.BaseUrl + path
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-TBA-Auth-Key", tbaAuthKey)
	return httpClient.Do(req)
}

func (client *TbaClient) postRequest(resource string, action string, body []byte) (*http.Response, error) {
	path := fmt.Sprintf("/api/trusted/v1/event/%s/%s/%s", client.eventCode, resource, action)
	signature := fmt.Sprintf("%x", md5.Sum(append([]byte(client.secret+path), body...)))

	httpClient := &http.Client{}
	request, err := http.NewRequest("POST", client.BaseUrl+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("X-TBA-Auth-Id", client.secretId)
	request.Header.Add("X-TBA-Auth-Sig", signature)
	response, err := httpClient.Do(request)
	if client.BaseUrl == tbaBaseUrl && err == nil && response.StatusCode == 200 {
		pingRequest, _ := http.NewRequest(
			"POST", fmt.Sprintf("https://cheesyarena.com/events/%s/%s", client.eventCode, resource), nil,
		)
		_, _ = httpClient.Do(pingRequest)
	}
	return response, err
}

func createTbaAlliance(teamIds [3]int, surrogates [3]bool, score *int, cards map[string]string) *TbaAlliance {
	alliance := TbaAlliance{Teams: []string{}, Surrogates: []string{}, Dqs: []string{}, Score: score}
	for i, teamId := range teamIds {
		if teamId == 0 {
			continue
		}
		teamKey := getTbaTeam(teamId)
		alliance.Teams = append(alliance.Teams, teamKey)
		if surrogates[i] {
			alliance.Surrogates = append(alliance.Surrogates, teamKey)
		}
		if cards != nil {
			if card, ok := cards[strconv.Itoa(teamId)]; ok && card == "red" {
				alliance.Dqs = append(alliance.Dqs, teamKey)
			}
		}
	}

	return &alliance
}

func createTbaScoringBreakdown(
	eventSettings *model.EventSettings,
	match *model.Match,
	matchResult *model.MatchResult,
	alliance string,
) map[string]any {
	var breakdown TbaScoreBreakdown
	var score *game.Score
	var scoreSummary *game.ScoreSummary
	if alliance == "red" {
		score = matchResult.RedScore
		scoreSummary = matchResult.RedScoreSummary()
	} else {
		score = matchResult.BlueScore
		scoreSummary = matchResult.BlueScoreSummary()
	}

	// 2026 REBUILT Fields Mapping

	// Adjust
	breakdown.AdjustPoints = 0

	// Auto
	breakdown.AutoTowerRobot1 = autoTowerMapping[score.AutoTowerLevel1[0]]
	breakdown.AutoTowerRobot2 = autoTowerMapping[score.AutoTowerLevel1[1]]
	breakdown.AutoTowerRobot3 = autoTowerMapping[score.AutoTowerLevel1[2]]
	breakdown.AutoTowerPoints = scoreSummary.AutoTowerPoints
	breakdown.TotalAutoPoints = scoreSummary.AutoPoints

	// Teleop
	breakdown.TotalTeleopPoints = scoreSummary.MatchPoints - scoreSummary.AutoPoints

	// Endgame
	breakdown.EndGameTowerRobot1 = endGameStatusMapping[score.EndgameStatuses[0]]
	breakdown.EndGameTowerRobot2 = endGameStatusMapping[score.EndgameStatuses[1]]
	breakdown.EndGameTowerRobot3 = endGameStatusMapping[score.EndgameStatuses[2]]
	breakdown.EndGameTowerPoints = scoreSummary.EndgameTowerPoints

	// Totals
	breakdown.TotalTowerPoints = scoreSummary.TotalTowerPoints
	breakdown.TotalPoints = scoreSummary.Score

	// 2026 Hub Score Details
	breakdown.HubScore.AutoCount = score.AutoFuelCount
	breakdown.HubScore.AutoPoints = scoreSummary.AutoFuelPoints
	breakdown.HubScore.EndgameCount = score.EndgameFuelCount
	breakdown.HubScore.EndgamePoints = score.EndgameFuelCount * 1
	breakdown.HubScore.Shift1Count = score.Shift1FuelCount
	breakdown.HubScore.Shift1Points = score.Shift1FuelCount * 1
	breakdown.HubScore.Shift2Count = score.Shift2FuelCount
	breakdown.HubScore.Shift2Points = score.Shift2FuelCount * 1
	breakdown.HubScore.Shift3Count = score.Shift3FuelCount
	breakdown.HubScore.Shift3Points = score.Shift3FuelCount * 1
	breakdown.HubScore.Shift4Count = score.Shift4FuelCount
	breakdown.HubScore.Shift4Points = score.Shift4FuelCount * 1
	breakdown.HubScore.TransitionCount = score.TransitionFuelCount
	breakdown.HubScore.TransitionPoints = score.TransitionFuelCount * 1
	breakdown.HubScore.TeleopCount = score.TeleopFuelCount
	breakdown.HubScore.TeleopPoints = scoreSummary.TeleopFuelPoints
	breakdown.HubScore.TotalCount = score.AutoFuelCount + score.TeleopFuelCount
	breakdown.HubScore.TotalPoints = scoreSummary.TotalFuelPoints

	// Fouls
	for _, foul := range score.Fouls {
		if foul.IsMajor {
			breakdown.MajorFoulCount++
		} else if foul.PointValue() > 0 {
			breakdown.MinorFoulCount++
		}
		if foul.Rule() != nil && foul.Rule().RuleNumber == "G206" {
			breakdown.G206Penalty = true
		}
	}
	breakdown.FoulPoints = scoreSummary.FoulPoints

	// Ranking Points & Achievements
	if match.ShouldUpdateRankings() {
		breakdown.EnergizedAchieved = scoreSummary.EnergizedRankingPoint
		breakdown.SuperchargedAchieved = scoreSummary.SuperchargedRankingPoint
		breakdown.TraversalAchieved = scoreSummary.TraversalRankingPoint
		breakdown.RP = scoreSummary.BonusRankingPoints

		if match.Status == game.RedWonMatch && alliance == "red" {
			breakdown.RP += 3
		} else if match.Status == game.BlueWonMatch && alliance == "blue" {
			breakdown.RP += 3
		} else if match.Status == game.TieMatch {
			breakdown.RP += 1
		}
	}

	breakdownMap := make(map[string]any)
	_ = mapstructure.Decode(breakdown, &breakdownMap)
	return breakdownMap
}
