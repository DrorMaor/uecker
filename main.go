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
var bti int = 0  // Batting Team Index (too long to write full name each time)
var Inning inning
var Count count
// we allow up to 5 errors per game (both teams combined)
// randomly, an error will occur, until this # is reached
var MaxErrors float64 = math.Floor(GetRand()*5)
var ErrorCount float64 = 0

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
    city string
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
    fmt.Printf ("The %s %s will be hosting the %s %s today.\n", Teams[1].city, Teams[1].name, Teams[0].city, Teams[0].name)
    Inning.num = 1
    Inning.TopBottom = false
    Teams[0].AtBatNum = 0
    Teams[1].AtBatNum = 0
    StartInning()
}

func GameOver() {
    os.Exit(-1)
}

func StartInning() {
    s := "We go to the "
    if !Inning.TopBottom {
        s += "top"
    } else {
        s += "bottom"
    }
    s += " of the " + strconv.Itoa(Inning.num)
    switch Inning.num {
        case 1:
            s += "st"
        case 2:
            s += "nd"
        default:
            s += "rd"
    }
    if GetRand() <.5 {
        s += " inning"
    }
    if GetRand() <.5 {
        s += ". The"
    } else {
        s += ", and the"
    }
    s += " score is "
    if GetRand() <.5 {
        s += Teams[0].city + " " + strconv.Itoa(Teams[0].score) + " and "
        s += Teams[1].city + " " + strconv.Itoa(Teams[1].score)
    } else {
        s += "the " + Teams[0].name + " " + strconv.Itoa(Teams[0].score) + " and "
        s += "the " + Teams[1].name + " " + strconv.Itoa(Teams[1].score)
    }
    fmt.Println (s)

    Inning.outs = 0
    for {
        DoAtBat()
        if Inning.outs == 3 {
            EndInning()
            break
        }
    }
}

func DoAtBat() {
    s := "Leading off"
    if Inning.outs >0 {
        s = "Batting next"
    }
    s += " for the " + Teams[bti].name
    if GetRand() <.5 {
        s += " will be "
    } else {
        s += " is "
    }
    s += Teams[bti].batters[Teams[bti].AtBatNum].name
    fmt.Println (s)

    Count.balls = 0
    Count.strikes = 0
    for {
        if (Count.balls < 4 && Count.strikes < 3) {
            DoPitch()
        } else {
            if Inning.outs == 3 {
                EndInning()
            }
        }
    }
}

func EndInning() {
    Inning.outs = 0
    if (Inning.TopBottom == true) {
        Inning.num ++
    }
    Inning.TopBottom = !Inning.TopBottom
    if Inning.TopBottom {
        bti = 0
    } else {
        bti = 1
    }

    //if (Inning.num < 9 || (Inning.num >=9 && Teams[0].score >= Teams[1].score) ) {
    if (Inning.num <9) {
        StartInning()
    } else {
        GameOver()
    }
}

func DoPitch() {
    p := GetRand()
    if p < .333 {
        // ball
        Count.balls ++
        if Count.balls == 4 {
            // walk
            Count.balls = 0
            AdvanceRunners(0, -1)
            DoAtBat()
        }
    } else if p >= .333 && p < .667  {
        // strike
        if ! (Count.strikes == 2 && GetRand() < .5) {
            // only add strike if it's not Strike 2 now and it's not a foul ball
            Count.strikes ++
            if Count.strikes == 3 {
                DoOut(true)
                DoAtBat()
            }
        }
    } else {
        // hit in play
        // determine if it's a hit or out
        h := GetRand()
        if h < Teams[bti].batters[Teams[bti].AtBatNum].AVG {
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
    AdvanceLineup()
}

func AdvanceLineup() {
    Teams[bti].AtBatNum ++
    if Teams[bti].AtBatNum % 9 == 0 {
        Teams[bti].AtBatNum = 0
    }
}

func AdvanceRunners(bases int, pos float64 ) {
    // bases: # of bases of hit
    // pos: defensive position where ball was hit
    switch bases {
        case -1: // out (sac fly)
            if pos >=6 && Inning.third {
                Inning.third = false  // the other 2 baserunners stay the same
                Teams[bti].score ++
            }
        case 0: // walk
            switch BasesStatus() {
                case "false|false|false":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = false
                case "true|false|false":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = false
                case "false|true|false":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = false
                case "false|false|true":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = true
                case "true|true|false":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = true
                case "true|false|true":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = true
                case "false|true|true":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = true
                case "true|true|true":
                    Teams[bti].score ++
                    Inning.first = true
                    Inning.second = true
                    Inning.third = true
            }
        case 1:
            switch BasesStatus() {
                case "false|false|false":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = false
                case "true|false|false":
                    Inning.first = true
                    if pos == 8 {
                        Inning.second = false
                        Inning.third = true
                    } else {
                        Inning.second = true
                        Inning.third = false
                    }
                case "false|true|false":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = false
                    Teams[bti].score ++
                case "false|false|true":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = false
                    Teams[bti].score ++
                case "true|true|false":
                    Inning.first = true
                    if pos == 7 || pos == 8 {
                        Inning.second = false
                        Inning.third = false
                        Teams[bti].score ++
                    } else {
                        Inning.second = false
                        Inning.third = true
                    }
                case "true|false|true":
                    Inning.first = true
                    Inning.third = false
                    if pos == 8 {
                        Inning.third = true
                    } else {
                        Inning.second = true
                    }
                    Teams[bti].score ++
                case "false|true|true":
                    if pos == 7 || pos == 8 {
                        Teams[bti].score += 2
                        Inning.first = true
                        Inning.second = false
                        Inning.third = false
                    } else {
                        Teams[bti].score ++
                        Inning.first = true
                        Inning.second = false
                        Inning.third = true
                    }
                case "true|true|true":
                    if pos == 7 || pos == 8 {
                        Teams[bti].score += 2
                        Inning.first = true
                        Inning.second = false
                        Inning.third = false
                    } else {
                        Teams[bti].score ++
                        Inning.first = true
                        Inning.second = true
                        Inning.third = true
                    }
            }
        case 2:
            switch BasesStatus() {
                case "false|false|false":
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "true|false|false":
                    Teams[bti].score ++
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "false|true|false":
                    Teams[bti].score ++
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "false|false|true":
                    Teams[bti].score ++
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "true|true|false":
                    Teams[bti].score += 2
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "true|false|true":
                    Teams[bti].score += 2
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "false|true|true":
                    Teams[bti].score += 2
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "true|true|true":
                    Teams[bti].score += 3
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
            }
        case 3:
            switch BasesStatus() {
                case "false|false|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "true|false|false":
                    Teams[bti].score ++
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "false|true|false":
                    Teams[bti].score ++
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "false|false|true":
                    Teams[bti].score ++
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "true|true|false":
                    Teams[bti].score += 2
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "true|false|true":
                    Teams[bti].score += 2
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "false|true|true":
                    Teams[bti].score += 2
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "true|true|true":
                    Teams[bti].score += 3
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
            }
        case 4:
            switch BasesStatus() {
                case "false|false|false":
                    Teams[bti].score ++
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "true|false|false":
                    Teams[bti].score += 2
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "false|true|false":
                    Teams[bti].score += 2
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "false|false|true":
                    Teams[bti].score += 2
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "true|true|false":
                    Teams[bti].score += 3
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "true|false|true":
                    Teams[bti].score += 3
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "false|true|true":
                    Teams[bti].score += 3
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "true|true|true":
                    Teams[bti].score += 4
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
            }
    }
}

func BasesStatus() string {
    return strconv.FormatBool(Inning.first) + "|" + strconv.FormatBool(Inning.second)  + "|" + strconv.FormatBool(Inning.third)
}

func DoOut(strikeout bool) {
    if !strikeout {
        pos := math.Floor(GetRand()*9) +1
        if pos >=7 {
            AdvanceRunners(-1, pos)
        }
    }
    Count.strikes = 0
    Count.balls = 0
    Inning.outs ++
    AdvanceLineup()
    if Inning.outs == 3 {
        EndInning()
    }
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
            case "city":
                Team.city = line[1]
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
