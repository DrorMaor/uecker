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
var RepeatChars int = 40  // for inning and other --------- separator

type inning struct {
    num int
    TopBottom bool  // top True, bottom False
    outs int
    LeadOff bool    // for the script, to say if it's the leadoff batter
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
	  FirstName string
    LastName string
    AVG float64
}

type pitcher struct {
    position string
	  FirstName string
    LastName string
    ERA float64
}

func main() {
    Teams[0] = GetLineup("AwayTeam", Teams[0])
    Teams[1] = GetLineup("HomeTeam", Teams[1])
    PlayBall()
}

func PlayBall() {
    GameScript(1)
    Inning.num = 1
    Inning.TopBottom = false
    Teams[0].AtBatNum = 0
    Teams[1].AtBatNum = 0
    StartInning()
}

func GameOver() {
    GameScript(7)
    os.Exit(-1)
}

func StartInning() {
    Inning.LeadOff = true
    Inning.outs = 0
    Inning.first = false
    Inning.second = false
    Inning.third = false
    if Inning.TopBottom {
        Inning.num ++
    }
    Inning.TopBottom = !Inning.TopBottom
    if Inning.TopBottom {
        bti = 0
    } else {
        bti = 1
    }

    for {
        DoAtBat()
        if Inning.outs == 3 {
            EndInning()
            break
        }
    }
}

func DoAtBat() {
    GameScript(3)
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
    // all this occurs at the end of an inning, BEFORE we increment the inning #
    GameScript(2)
    // first 8 innings, always start the next inning (frame)
    if Inning.num < 9  {
        StartInning()
    }
    // 9th inning, end of top half
    if Inning.num == 9 && !Inning.TopBottom {
        if Teams[1].score > Teams[0].score {
            // no need for frame, home team won
            GameOver()
        } else {
            // do top of the 9th
            StartInning()
        }
    }

    // bottom of the 9th
    if Inning.num >= 9 && Inning.TopBottom {
        if Teams[1].score != Teams[0].score {
            // game over after 9 innings
            GameOver()
        } else {
            // do extra innings
            StartInning()
        }
    }
}

func DoPitch() {
    if GetRand() <.75 {
        // ball or strike
        if GetRand() < .5 {
            // ball
            Count.balls ++
            GameScript(12)
            if Count.balls == 4 {
                // walk
                Count.balls = 0
                AdvanceRunners(0, -1)
                DoAtBat()
            }
        } else {
            // strike
            s := GetRand()
            if ! (Count.strikes == 2 && s >=.667) {
                // only add strike if it's not Strike 2 now and it's not a foul ball
                Count.strikes ++
                if s <.333 {
                    GameScript(13)
                } else if s >=.333 && s <.667 {
                    GameScript(14)
                } else {
                    GameScript(15)
                }

                if Count.strikes == 3 {
                    DoOut(true)
                    DoAtBat()
                }
            }
        }
    } else {
        // hit in play
        // determine if it's a hit or out
        if GetRand() < Teams[bti].batters[Teams[bti].AtBatNum].AVG {
            // he's on base
            // determine which hit type
            r := GetRand()
            if r < .1 {
                DoHit(4)
                GameScript(8)
            } else if r >= .1 && r < .15 {
                DoHit(3)
                GameScript(9)
            } else if r >= .15 && r < .33 {
                DoHit(2)
                GameScript(10)
            } else {
                DoHit(1)
                GameScript(11)
            }
            GameScript(6)
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
    GameScript(6)
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
    } else {
        DoAtBat()
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
        stat, err := strconv.ParseFloat(line[3], 64)
        if err != nil {
          // insert error handling here
        }
        switch line[0] {
            case "SP":
        		fallthrough
            case "RP":
        		Team.pitchers = append(Team.pitchers, pitcher{line[0], line[1], line[2], stat})
            case "city":
                Team.city = line[1]
        	  case "name":
                Team.name = line[1]
            case "WL":
                W, err := strconv.ParseFloat(line[1], 64)
                if err != nil {
                  // insert error handling here
                }
                L, err := strconv.ParseFloat(line[2], 64)
                if err != nil {
                  // insert error handling here
                }
                Team.pct = W / (W + L)
        	default:  // regular position players
        		Team.batters = append(Team.batters, batter{line[0], line[1], line[2], stat})
        }
	}
    return Team
}

func GameScript(id int) {
    var script string = ""
    switch id {
        case 1:
            // start of game
            script += fmt.Sprintf("The %s %s will be hosting the %s %s\n\n", Teams[1].city, Teams[1].name, Teams[0].city, Teams[0].name)
            // print the lineup
            for _, team := range Teams {
                script += fmt.Sprintf("%s's lineup:\n", team.city)
                for _, batter := range team.batters {
                    script += fmt.Sprintf("%s, %s %s\n", batter.position, batter.FirstName, batter.LastName)
                }
                script += "\n"
            }
        case 2:
            // end of inning
            script += "End of "
            if !Inning.TopBottom {
                script += "top"
            } else {
                script += "bottom"
            }
            script += fmt.Sprintf(" of %d", Inning.num)
            switch Inning.num {
                case 1:
                    script += "st"
                case 2:
                    script += "nd"
                case 3:
                    script += "rd"
                default:
                    script += "th"
            }
            script += ". " + ScoreScript()
            script += "\n" + strings.Repeat("-", RepeatChars) + "\n" + strings.Repeat("-", RepeatChars) + "\n\n"
        case 3:
            // each at bat
            script += strings.Repeat("-", RepeatChars) + "\n"
            if !Inning.LeadOff {
                script += "Batting"
            } else {
                script += "Leading off"
            }
            script += fmt.Sprintf(" for %s, %s %s %s", Teams[bti].city, Teams[bti].batters[Teams[bti].AtBatNum].position, Teams[bti].batters[Teams[bti].AtBatNum].FirstName, Teams[bti].batters[Teams[bti].AtBatNum].LastName)

            Inning.LeadOff = false
        case 6:
            // after runners advancing
            switch BasesStatus() {
                case "true|false|false":
                    script += "A runner at first base"
                case "false|true|false":
                    script += "A runner at second base"
                case "false|false|true":
                    script += "A runner at third base"
                case "true|true|false":
                    script += "Runners at first and second"
                case "true|false|true":
                    script += "Runners at first and third"
                case "false|true|true":
                    script += "Runners at second and third"
                case "true|true|true":
                    script += "Bases loaded"
            }
        case 7:
            // game over
            script += "Game over. Final " + ScoreScript()
        case 8:
            // home run
            script += "Homerun to " + RandomField()
        case 9:
            // triple
            script += "Triple to " + RandomField()
        case 10:
            // double
            script += "Double to " + RandomField()
        case 11:
            // single
            script += "Single to " + RandomField()
        case 12:
            // ball
            script += "Ball. " + CountScript()
        case 13:
            // swinging strike
            script += "Swing and a miss. " + CountScript()
        case 14:
            // called strike
            script += "Called strike. " + CountScript()
        case 15:
            // foul ball
            script += "Foul ball. " + CountScript()
        case 16:

    }
    fmt.Println(script)
}

func ScoreScript() string {
    return fmt.Sprintf("Score: %s %d, %s %d", Teams[0].city, Teams[0].score, Teams[1].city, Teams[1].score)
}

func CountScript() string {
    if Count.balls == 4{
        return "Walk"
    } else if Count.strikes == 3 {
        return "Strikeout"
    } else {
        return fmt.Sprintf("Count %d-%d", Count.balls, Count.strikes)
    }
}

func RandomField () string {
    var field string = ""
    r := GetRand()
    if r <.2 {
        field = "left field"
    } else if r >=.2 && r <.4 {
        field = "left center"
    } else if r >=.4 && r <.6 {
        field = "center field"
    } else if r >=.6 && r <.8 {
        field = "right center"
    } else {
        field = "right field"
    }
    return field
}
