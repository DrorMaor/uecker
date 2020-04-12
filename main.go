package main

import (
    "fmt"
    "log"
    "bufio"
    "os"
    "strings"
    "strconv"
)

var AwayTeam team
var HomeTeam team

type inning struct {
    num int
    TopBottom bool  // top True, bottom False
    // runners on bases:
    first bool
    second bool
    third bool
}

type team struct {
    name string
    pct float64
    batters []batter
    pitchers []pitcher
    batter int  // # in the lineup
    pitcher int // # in the pitchers list
    score int
}

type batter struct {
    position string
	name string
    AVG float64
}
type pitcher struct {
    position string
    name string
    ERA float64
}

func main() {
    AwayTeam = GetLineup("AwayTeam", AwayTeam)
    HomeTeam = GetLineup("HomeTeam", HomeTeam)

    fmt.Printf("%+v\n\n", AwayTeam)
    fmt.Printf("%+v\n\n", HomeTeam)
}

func GetLineup(FileName string, Team team ) team {
    file, err := os.Open(FileName)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}
	file.Close()

	for _, eachline := range txtlines {
        line := strings.Split(eachline, "|")
        stat, err := strconv.ParseFloat(line[2], 64)
        if err != nil {
          // insert error handling here
        }
        switch line[0] {
            case "SP":
        		Team.pitchers = append(Team.pitchers, pitcher{line[0], line[1], stat})
            case "RP":
        		Team.pitchers = append(Team.pitchers, pitcher{line[0], line[1], stat})
        	case "name":
                Team.name = line[1]
                WL := strings.Split(line[2], "-")
                W, err := strconv.ParseFloat(WL[0], 64)
                if err != nil {
                  // insert error handling here
                }
                L, err := strconv.ParseFloat(WL[1], 64)
                if err != nil {
                  // insert error handling here
                }
                Team.pct = W / (W + L)
        	default:
        		Team.batters = append(Team.batters, batter{line[0], line[1], stat})
        }
	}
    return Team
}
