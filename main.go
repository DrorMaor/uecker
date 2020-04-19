package main

import (
    "fmt"
    "log"
    "bufio"
    "os"
    "strings"
    "strconv"
    "math"
    "math/rand"
    "time"
)

var Teams [2]team
var BattingTeamIndex int = 0
var Inning inning
var Count count
// we allow up to 5 errors per game (both teams combined)
// randomly, an error will occur, until this # is reached
var MaxErrors int = math.Floor(GetRand()*5)
var ErrorCount int = 0

type inning struct {
    num int
    TopBottom bool  // top True, bottom False
    outs int
    // runners on bases:
    first bool
    second bool
    third bool
}

type count struct {
    balls int
    strikes int
}

type team struct {
    name string
    pct float64
    batters []batter
    pitchers []pitcher
    batter int  // # in the lineup
    pitcher int // # in the pitchers list
    score int
    AtBatNum int
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
    Teams[0] = GetLineup("AwayTeam", Teams[0])
    Teams[1] = GetLineup("HomeTeam", Teams[1])
    PlayBall()
}

func PlayBall() {
    Inning.num = 1
    Inning.TopBottom = false
    for {
        if (Inning.num < 9 || (Inning.num >=9 && Team[0].score >= Team[1].score) ) {
            DoInning()
        } else {
            GameOver()
        }
    }
}

func GameOver() {

}

func DoInning() {
    Inning.outs = 0
    for {
        if (Inning.outs == 3) {
            if (Inning.TopBottom == true) {
                Inning.num ++
                Inning.TopBottom = !Inning.TopBottom
                if Inning.TopBottom {
                    BattingTeamIndex = 0
                } else {
                    BattingTeamIndex = 1
                }
            }
            break
        } else {
            AtBat()
        }
    }
}

func AtBat() {
    Count.balls = 0
    Count.strikes = 0
    for {
        if (Count.balls < 4 && Count.strikes < 3) {
            Pitch()
        } else {
            break
        }
    }
}

func Pitch() {
    p := GetRand()
    if r < .333 {
        // ball
        Count.balls ++
        if Count.balls == 4 {
            // walk
            AdvanceRunners(0, -1)
        }
    } else if p >= .333 && r < .667  {
        // strike
        if ! (Count.strikes == 2 && GetRand() < .5) {
            // only add strike if it's not Strike 2 now and it's not a foul ball
            Count.strikes ++
            if Count.strikes == 3 {
                DoOut(true)
            }
        }
    } else {
        // hit in play
        // determine if it's a hit or out
        h := GetRand()
        if h < team.batters[team.AtBatNum].AVG {
            // he's on base
            // determine which hit type
            r := GetRand()
            if r < .1 {
                DoHit(4)
            } else if r >= .1 && r < .15 {
                DoHit(3)
            } else if r >= .15 && r < .33 {
                DoHit(2)
            } else {
                DoHit(1)
            }
        } else {
            // he's out
            DoOut(false)
        }
    }
}

func DoHit(bases int) {
    // most hits are out of the infield, so we assume them here
    outfield := math.Floor(GetRand()*3) + 6  // left, center, or right field (nfk"m for runner scoring from second)
    AdvanceRunners(bases, outfield)
}

func AdvanceRunners(bases int, pos int ) {
    // bases: # of bases of hit
    // pos: defensive position where ball was hit
    switch bases {
        case -1: // out (sac fly)
            if pos >=6 && Inning.third {
                Inning.third = false
                Teams[BattingTeamIndex].score ++
            }
        case 0: // walk
            switch BasesStatus() {
                case "false|false|false":
                    Inning.first = true
                case "true|false|false":
                    Inning.second = true
                case "false|true|false":
                    Inning.first = true
                case "false|false|true":
                    Inning.first = true
                case "true|true|false":
                    Inning.third = true
                case "true|true|false":
                    Inning.second = true
                case "true|false|true":
                    Inning.first = true
                case "true|true|true":
                    Teams[BattingTeamIndex].score ++
            }
        case 1:
            switch BasesStatus() {
                case "false|false|false":
                    Inning.first = true
                case "true|false|false":
                    Inning.first = true
                    if pos == 8 {
                        Inning.third = true
                    } else {
                        Inning.second = true
                    }
                case "false|true|false":
                    Inning.first = true
                    Inning.second = false
                    Teams[BattingTeamIndex].score ++
                case "false|false|true":
                    Inning.first = true
                    Teams[BattingTeamIndex].score ++
                    Inning.first = true
                case "true|true|false":

                case "true|true|false":

                case "true|false|true":

                case "true|true|true":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                    Teams[BattingTeamIndex].score += 4
            }
        case 2:

        case 3:

        case 4:

    }
}

func BasesStatus () {
    return strconv.FormatBool(Inning.first) + "|" + strconv.FormatBool(Inning.second)  + "|" + strconv.FormatBool(Inning.third)
}

func DoOut(strikeout bool) {
    if !strikeout {
        pos := math.Floor(GetRand()*9) +1
        if pos >=7 {
            AdvanceRunners(-1, pos)
        }
    }
    Inning.outs ++
}

func GetRand() float64 {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    return r.Float64()
}

func GetLineup(FileName string, Team team ) team {
    file, err := os.Open(FileName)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtLines []string
	for scanner.Scan() {
		txtLines = append(txtLines, scanner.Text())
	}
	file.Close()

	for _, eachline := range txtLines {
        line := strings.Split(eachline, "|")
        stat, err := strconv.ParseFloat(line[2], 64)
        if err != nil {
          // insert error handling here
        }
        switch line[0] {
            case "SP":
        		fallthrough
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
        	default:  // regular position players
        		Team.batters = append(Team.batters, batter{line[0], line[1], stat})
        }
	}
    return Team
}
