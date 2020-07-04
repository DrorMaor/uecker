package main

import (
    "fmt"
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
var MaxErrors float64 = math.Floor(GetRand()*6)
var ErrorCount int = 0

// for inning and other --------- separator
var RepeatChars int = 40

// this is where we save the ongoing play-by-play calls
var FullGameScript string = ""
var PlayNum int = 0 // for debugging

// this will increment by 1 with each top/bottom of inning, and will be a running total, but we can determine (for printing reasons) the actual inning # and top/bottom
var InningFrame int = -1
// this will be translated from the InningFrame val (for script purposes)
var InningNum int = 0

type inning struct {
    outs int
    LeadOff bool    // for the script, to say if it's the leadoff batter
    // runners on bases
    runners [3]bool
}

type count struct {
    balls int
    strikes int
}

type team struct {
    city string
    name string
    short string  // 3 letter short team (such as MIL)
    pct float64   // winning percentage: W/(W+L)
    batters []batter
    pitchers []pitcher
    pitcher int  // # in the pitchers list
    score int
    AtBatNum int
    Boxscore boxscore
    CurPitcherInns int  // this will help us determine when a pitching change is needed
}

type batter struct {
    pos string
	FirstName string
    LastName string
    BatterHitsPct batterHitsPct  // refer to comments in that struct
    AVG float64
}

type batterHitsPct struct {
    // this %age will be used for the random determining what hit they get
    // (it's computed from the # of DBL/TPL/HR's they hit divided by # of AtBats)
    DBL float64
    TPL float64
    HR float64
}

type pitcher struct {
    pos string
	FirstName string
    LastName string
    ERA float64
    AvgInnPerGame int
    RS string   // Reliever or Starter
}

type boxscore struct {
    inn []int  // will store the # of runs per each frame
    H int
    E int
}

func main() {
    Teams[0] = GetLineup("AwayTeam", Teams[0])
    Teams[1] = GetLineup("HomeTeam", Teams[1])
    PlayBall()
}

func PlayBall() {
    GameScript(1, "")
    Teams[0].AtBatNum = 0
    Teams[1].AtBatNum = 0
    StartInning()
}

func GameOver() {
    GameScript(7, "")
    DrawBoxscore()
    f, _ := os.Create("GameScript")
    f.WriteString(FullGameScript)
    f.Close()
    os.Exit(-1)
}

func DrawBoxscore() {
    box := "\n"
    // this is the heading of the boxscore
    box += "    "
    inns := len(Teams[0].Boxscore.inn)
    for i := 1; i <= inns; i++ {
        box += fmt.Sprintf("%d ", i)
    }
    box += " R H E \n"
    // now both teams' #s
    for _, team := range Teams {
        box += team.short + " "
        if len(team.Boxscore.inn) < inns {
            team.Boxscore.inn = append(team.Boxscore.inn, -1)
        }
        for i := 0; i < inns; i++ {
            if team.Boxscore.inn[i] == -1 {
                box += "- "
            } else {
                box += fmt.Sprintf("%d ", team.Boxscore.inn[i])
            }
        }
        box += fmt.Sprintf(" %d %d %d\n", team.score, team.Boxscore.H, team.Boxscore.E)
    }
    FullGameScript += box + "\n"

    fmt.Println (box)

    f, _ := os.OpenFile("boxscore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    f.WriteString(box)
    f.Close()
}

func StartInning() {
    // reset the inning numbers
    Inning.outs = 0
    Inning.LeadOff = true
    SetRunnersStatus([3]bool {false, false, false})

    CheckPitchingChange()

    InningFrame ++
    InningNum = int(math.Floor(float64(InningFrame) / 2) + 1)
    bti = InningFrame % 2
    Teams[bti].Boxscore.inn = append(Teams[bti].Boxscore.inn, 0)
    for {
        DoAtBat()
        if Inning.outs == 3 {
            EndInning()
            break
        }
    }
}

func CheckPitchingChange() {
    if Teams[bti].pitchers[Teams[bti].pitcher].AvgInnPerGame == Teams[bti].CurPitcherInns {
        if len(Teams[bti].pitchers) > Teams[bti].pitcher + 1 {
            Teams[bti].pitcher ++
            Teams[bti].CurPitcherInns = 0
        }
    }
}

func DoAtBat() {
    GameScript(3, "")
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
    // all this occurs at the end of an inning, BEFORE we increment the inning # and frame
    // (which we do at the beginning of StartInning)

    Teams[bti].CurPitcherInns ++  // needed to determine if pitching change is due

    // determine whether to end the game, or start another inning
    if InningFrame == 16 && Teams[1].score > Teams[0].score {
        // bottom of the 9th, home team ahead
        GameOver()
    } else if InningFrame >= 17 && bti == 1 && Teams[1].score != Teams[0].score {
        // extra innings, bottom of frame, any team ahead
        GameOver()
    } else {
        // 1-8 innings, or any other extra inning
        GameScript(2, "")
        StartInning()
    }
}

func DoPitch() {
    // assuming that an average count of an at-bat is 2-1 and then he hits it, so that's a .25 chance of him hitting it
    if GetRand() < 0.75 {
        // ball or strike (2/3 chance of a ball)
        if GetRand() < 0.667 {
            // ball
            Count.balls ++
            GameScript(12, "")
            if Count.balls == 4 {
                // walk
                AdvanceRunners(0, -1)
                AdvanceLineup()
                DoAtBat()
            }
        } else {
            // strike
            s := GetRand()
            if ! (Count.strikes == 2 && s >= 0.667) {
                // only add strike if it's not Strike 2 now and it's not a foul ball
                Count.strikes ++
                if s < 0.333 {
                    GameScript(13, "Swing and a miss")
                } else if s >= 0.333 && s < 0.667 {
                    tCorner := ""
                    if GetRand() < 0.5 {
                        tCorner = "inside"
                    } else {
                        tCorner = "outside"
                    }
                    GameScript(13, "Called strike on the " + tCorner + " corner")
                } else {
                    f := GetRand()
                    fbText := "Foul ball "
                    if f < 0.2 {
                        fbText += "behind the plate"
                    } else if f >= 0.2 && f < 0.4 {
                        fbText += "third base side"
                    } else if f >= 0.4 && f < 0.6 {
                        fbText += "first base side"
                    } else if f >= 0.6 && f < 0.8 {
                        fbText += "to left field"
                    } else  {
                        fbText += "to right field"
                    }
                    GameScript(13, fbText)
                }

                if Count.strikes == 3 {
                    DoOut(true)
                }
            }
        }
    } else {
        // hit in play

        // determine if it's a hit or out
        CurrBtr := Teams[bti].batters[Teams[bti].AtBatNum]
        // ERA3 is ERA adjusted to 3.0
        // (3.00 is a decent ERA, and anothing higher would make the batter stronger,
        //  and anything lower would make the batter weaker)
        ERA3 := Teams[bti].pitchers[Teams[bti].pitcher].ERA - 3.33
        // GBOP = Getting Batter Out Percentage, we adjust the ERA3 based on the AVG
        // so we have a fair chance at a hit/out, based on both pitcher & batter
        GBOP := CurrBtr.AVG + (ERA3 / 50)
        if GetRand() < GBOP {
            // he's on base
            // determine which hit type (param is # of bases in hit)
            r := GetRand()
            if r < CurrBtr.BatterHitsPct.HR {
                DoHit(4, "Homerun to " + RandomField())
            } else if r >= CurrBtr.BatterHitsPct.HR && r < (CurrBtr.BatterHitsPct.HR + CurrBtr.BatterHitsPct.TPL) {
                // triple (can't use RandomField here, because a triple will rarely NOT be in right field)
                tText := "Triple to "
                if GetRand() < 0.5 {
                    tText += "right center"
                } else {
                    tText += "right field"
                }
                DoHit(3, tText)
            } else if r >= (CurrBtr.BatterHitsPct.HR + CurrBtr.BatterHitsPct.TPL) && r < (CurrBtr.BatterHitsPct.HR + CurrBtr.BatterHitsPct.TPL + CurrBtr.BatterHitsPct.DBL) {
                DoHit(2, "Double to " + RandomField())
            } else {
                DoHit(1, "Single to " + RandomField())
            }
        } else {
            if !TryError() {
                // he's out
                DoOut(false)
            }
        }
    }
}

func TryError() bool {
    error := false
    if ErrorCount < int(MaxErrors) {
        // try throwing an error
        // (this is based on 80 atbats per game: 27 min per team, plus average 3 walks and 10 hits)
        if GetRand() < (MaxErrors / 80) {
            GameScript(18, "")
            error = true
            // it's the NON batting team that gets charged with the error
            if bti == 0 {
                Teams[1].Boxscore.E ++
            } else {
                Teams[0].Boxscore.E ++
            }

            ErrorCount ++
            AdvanceRunners(-2, -1)
            AdvanceLineup()
            DoAtBat()
        }
    }
    return error
}

func DoHit(bases int, text string) {
    GameScript(8, text)

    // most base hits are out of the infield, so we assume them here
    outfield := int(math.Floor(GetRand()*3)) + 7  // left, center, or right field (nfk"m for runner scoring from second)
    AdvanceRunners(bases, outfield)
    AdvanceLineup()
    Teams[bti].Boxscore.H ++
    DoAtBat()
}

func AdvanceLineup() {
    Teams[bti].AtBatNum ++
    if Teams[bti].AtBatNum % 9 == 0 {
        Teams[bti].AtBatNum = 0
    }
}

func SetRunnersStatus(runners [3]bool) {
    for i := 0; i <= 2; i++ {
        Inning.runners[i] = runners[i]
    }
}

func TryDoublePlay(pos int) string {
    var dbTurned bool = true // this will be the result (will be the default value here, unless it's set to false)

    // who's the middle man for the double play
    var players = [3]int {pos, 4, 3}
    players[1] = 4
    if pos == 5 {
        players[1] = 6
    }

    switch BasesStatus() {
        case "000":
            dbTurned = false
        case "100":
            Inning.runners[0] = false
        case "010":
            fallthrough
        case "001":
            dbTurned = false
        case "110":
            fallthrough
        case "101":
            SetRunnersStatus([3]bool {false, false, true})
        case "011":
            dbTurned = false
        case "111":
            SetRunnersStatus([3]bool {true, true, false})
            players = [3]int {pos, 2, 5}
    }

    var dpText string = ""
    if dbTurned {
        dpText = strconv.Itoa(players[0]) + "-" + strconv.Itoa(players[1]) + "-" + strconv.Itoa(players[2])
    }
    return dpText
}

func IncrementScore(runs int, HR bool) {
    walkoff := false
    if InningFrame >= 17 && bti == 1 && Teams[1].score + runs > Teams[0].score {
        // in the bottom of the 9+ inning, test for a walkoff, and if so, only count the # of runs needed to win
        // (unless it's a HR, then all runs count)
        walkoff = true
        if !HR {
            fmt.Println("walkoff + " + strconv.Itoa(runs))
            runs = Teams[0].score - Teams[1].score + 1
            fmt.Println("adjusted runs: " + strconv.Itoa(runs))
        }
    }
    Teams[bti].score += runs
    rText := ""
    switch (runs) {
        case 1:
            rText = "1 run scores"
        default:
            rText = strconv.Itoa(runs) + " runs score"
    }
    Teams[bti].Boxscore.inn[len(Teams[bti].Boxscore.inn)-1] += runs
    GameScript(17, rText)
    if walkoff {
        GameOver()
    }
}

func AdvanceRunners(bases int, pos int) {
    // bases: # of bases of hit
    // pos: defensive position where ball was hit (1 based)

    var DoGameScript bool = true  // if it's a flyout and not a sac fly, then no runners advanced, so nothing to print
    switch bases {
        case -2: // error (assumed one base advance per runner, plus batter safe at first)
            switch BasesStatus() {
                case "000":
                    SetRunnersStatus([3]bool {true, false, false})
                case "100":
                    SetRunnersStatus([3]bool {true, true, false})
                case "010":
                    SetRunnersStatus([3]bool {true, false, true})
                case "001":
                    SetRunnersStatus([3]bool {true, false, false})
                    IncrementScore(1, false)
                case "110":
                    SetRunnersStatus([3]bool {true, true, true})
                case "101":
                    SetRunnersStatus([3]bool {true, true, false})
                    IncrementScore(1, false)
                case "011":
                    SetRunnersStatus([3]bool {true, false, true})
                    IncrementScore(1, false)
                case "111":
                    IncrementScore(1, false)
            }
        case -1: // out (sac fly)
            if pos >= 7 && Inning.runners[2] && Inning.outs < 2 {
                Inning.runners[2] = false  // the other 2 baserunners stay the same
                IncrementScore(1, false)
            } else {
                DoGameScript = false
            }
        case 0: // walk
            switch BasesStatus() {
                case "000":
                    SetRunnersStatus([3]bool {true, false, false})
                case "100":
                    fallthrough
                case "010":
                    SetRunnersStatus([3]bool {true, true, false})
                case "001":
                    SetRunnersStatus([3]bool {true, false, true})
                case "110":
                    fallthrough
                case "101":
                    fallthrough
                case "011":
                    SetRunnersStatus([3]bool {true, true, true})
                case "111":
                    IncrementScore(1, false)
            }
        // from now on these are # of bases in the hit
        case 1:
            switch BasesStatus() {
                case "000":
                    SetRunnersStatus([3]bool {true, false, false})
                case "100":
                    if pos == 9 {
                        // runner will advance from 1st to 3rd on a single to right
                        SetRunnersStatus([3]bool {true, false, true})
                    } else {
                        SetRunnersStatus([3]bool {true, true, false})
                    }
                case "010":
                    // runner will score from second
                    fallthrough
                case "001":
                    SetRunnersStatus([3]bool {true, false, false})
                    IncrementScore(1, false)
                case "110":
                    if pos >= 8 {
                        SetRunnersStatus([3]bool {true, false, true})
                    } else {
                        SetRunnersStatus([3]bool {true, true, false})
                    }
                    IncrementScore(1, false)
                case "101":
                    if pos >= 8 {
                        SetRunnersStatus([3]bool {true, false, true})
                    } else {
                        SetRunnersStatus([3]bool {true, true, false})
                    }
                    IncrementScore(1, false)
                case "011":
                    if pos >= 8 {
                        SetRunnersStatus([3]bool {true, false, false})
                        IncrementScore(2, false)
                    } else {
                        SetRunnersStatus([3]bool {true, false, true})
                        IncrementScore(1, false)
                    }
                case "111":
                    if pos >= 8 {
                        SetRunnersStatus([3]bool {true, false, true})
                        IncrementScore(2, false)
                    } else {
                        SetRunnersStatus([3]bool {true, true, true})
                        IncrementScore(1, false)
                    }
            }
        case 2:
            switch BasesStatus() {
                case "000":
                case "100":
                    fallthrough
                case "010":
                    fallthrough
                case "001":
                    IncrementScore(1, false)
                case "110":
                    fallthrough
                case "101":
                    fallthrough
                case "011":
                    IncrementScore(2, false)
                case "111":
                    IncrementScore(3, false)
            }
            SetRunnersStatus([3]bool {false, true, false})  // will always clear the bases (besides for batter himself)
        case 3:
            switch BasesStatus() {
                case "000":
                case "100":
                    fallthrough
                case "010":
                    fallthrough
                case "001":
                    IncrementScore(1, false)
                case "110":
                    fallthrough
                case "101":
                    fallthrough
                case "011":
                    IncrementScore(2, false)
                case "111":
                    IncrementScore(3, false)
            }
            SetRunnersStatus([3]bool {false, false, true}) // will always clear the bases
        case 4:
            switch BasesStatus() {
                case "000":
                    IncrementScore(1, true)
                case "100":
                    fallthrough
                case "010":
                    fallthrough
                case "001":
                    IncrementScore(2, true)
                case "110":
                    fallthrough
                case "101":
                    fallthrough
                case "011":
                    IncrementScore(3, true)
                case "111":
                    IncrementScore(4, true)
            }
            SetRunnersStatus([3]bool {false, false, false}) // will always clear the bases
    }
    if DoGameScript {
        GameScript(6, "")
    }
}

func BasesStatus() string {
    var retVal string = ""
    for i := 0; i <= 2; i++ {
        if Inning.runners[i] {
            retVal += "1"
        } else {
            retVal += "0"
        }
    }
    return retVal
}

func DoOut(strikeout bool) {
    if !strikeout {
        r := GetRand()
        script := ""
        pos := 0
        // much less likelihood that the pitcher or catcher will do the putout, so we give them a smaller probabililty
        if r < 0.0625 {
            script = "Groundout to the pitcher"
            pos = 1
        } else if r >= 0.0625 && r < 0.125 {
            script = "Popout to the catcher"
            pos = 2
        } else if r >= 0.125 && r < 0.25 {
            script = "Groundout to first"
            pos = 3
        } else if r >= 0.25 && r < 0.375 {
            script = "Groundout to second"
            pos = 4
        } else if r >= 0.375 && r < 0.5 {
            script = "Groundout to third"
            pos = 5
        } else if r >= 0.5 && r < 0.625 {
            script = "Groundout to short"
            pos = 6
        } else if r >= 0.625 && r < 0.75 {
            script = "Flyout to left"
            pos = 7
        } else if r >= 0.75 && r < 0.875 {
            script = "Flyout to center"
            pos = 8
        } else {
            script = "Flyout to right"
            pos = 9
        }
        GameScript(14, script)

        if pos >= 7 {
            AdvanceRunners(-1, pos)
        } else {
            if Inning.outs < 2 {
                if pos != 2 {  // rare to have a catcher start a double play
                    dbText := TryDoublePlay(pos)
                    if len(dbText) > 0 {
                        // dbText returns true only means that it's possible for a DP,
                        // but still there's a 10% chance it won't be turned
                        if GetRand() < 0.9 {
                            dbText = "Double play " + dbText
                            Inning.outs ++  // this will only be the EXTRA out
                        } else {
                            dbText = "Attempted double play. Only 1 out recorded"
                        }
                        GameScript(16, dbText)
                    }
                    if Inning.outs < 2 {
                        // less than 2 outs will make sure that it won't print a runner on base on an inning ending double play
                        GameScript(6, "")  // baserunners
                    }
                }
            }
        }
    }

    Count.strikes = 0
    Count.balls = 0
    Inning.outs ++  // always increment the regular out

    AdvanceLineup()
    if Inning.outs == 3 {
        EndInning()
    } else {
        GameScript(15, "")
        DoAtBat()
    }
}

func GetRand() float64 {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    return r.Float64()
}

func GetLineup(FileName string, Team team) team {
    file, _ := os.Open(FileName)
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtLines []string
	for scanner.Scan() {
		txtLines = append(txtLines, scanner.Text())
	}
	file.Close()

    var RS string = ""   // Reliever or Starter
	for _, eachline := range txtLines {
        line := strings.Split(eachline, "|")
        switch line[0] {
            case "city":
                Team.city = line[1]
        	case "name":
                Team.name = line[1]
            case "short":
                Team.short = line[1]
            case "WL":
                W, _ := strconv.ParseFloat(line[1], 64)
                L, _ := strconv.ParseFloat(line[2], 64)
                Team.pct = W / (W + L)
            case "SP":
                RS = "S"
                fallthrough
            case "RP":
                RS = "R"
                ERA, _ := strconv.ParseFloat(line[3], 64)
                AvgInnPerGame, _ := strconv.Atoi(line[4])
                Team.pitchers = append(Team.pitchers, pitcher{line[0], line[1], line[2], ERA, AvgInnPerGame, RS})
        	default:  // regular position players
                Team.batters = append(Team.batters, AddBatterToLineup(line))
        }
	}
    return Team
}

func AddBatterToLineup(line []string) batter {
    pos := line[0]
    FirstName := line[1]
    LastName := line[2]
    H, _ := strconv.Atoi(line[3])
    DBL, _ := strconv.Atoi(line[4])
    TPL, _ := strconv.Atoi(line[5])
    HR, _ := strconv.Atoi(line[6])
    var BatterHitsPct batterHitsPct = batterHitsPct{float64(DBL) / float64(H), float64(TPL) / float64(H), float64(HR) / float64(H)}
    AVG, _ := strconv.ParseFloat(line[7], 64)
    var Batter batter = batter{pos, FirstName, LastName, BatterHitsPct, float64(AVG)}
    return Batter
}

func GameScript(id int, text string) {
    var script string = ""
    switch id {
        case 1:
            // start of game
            // print the lineups
            for _, team := range Teams {
                script += fmt.Sprintf("%s's lineup:\n", team.city)
                for _, batter := range team.batters {
                    script += fmt.Sprintf("%s, %s %s (AVG: %.3f)\n", batter.pos, batter.FirstName, batter.LastName, batter.AVG)
                }
                script += fmt.Sprintf("Starting pitcher: %s %s (ERA: %.2f)\n\n", team.pitchers[0].FirstName, team.pitchers[0].LastName, team.pitchers[0].ERA)
            }
        case 2:
            // end of inning
            script = "End of " + InningScript()
            script += ". " + ScoreScript()
            script += "\n" + strings.Repeat("-", RepeatChars) + "\n" + strings.Repeat("-", RepeatChars) + "\n\n"
        case 3:
            // each at bat
            script = strings.Repeat("-", RepeatChars) + "\n"
            if !Inning.LeadOff {
                script += "Batting"
            } else {
                script += "Leading off in the " + InningScript()
            }
            script += fmt.Sprintf(" for %s, %s %s %s", Teams[bti].short, Teams[bti].batters[Teams[bti].AtBatNum].pos, Teams[bti].batters[Teams[bti].AtBatNum].FirstName, Teams[bti].batters[Teams[bti].AtBatNum].LastName)

            Inning.LeadOff = false
        case 6:
            // after runners advancing
            switch BasesStatus() {
                case "000":
                    //script = "Bases empty"
                case "100":
                    script = "A runner at first base"
                case "010":
                    script = "A runner at second base"
                case "001":
                    script = "A runner at third base"
                case "110":
                    script = "Runners at first and second"
                case "101":
                    if GetRand() < 0.5 {
                        script = "Runners at first and third"
                    } else {
                        script = "Runners at the corners"
                    }
                case "011":
                    script = "Runners at second and third"
                case "111":
                    script = "Bases loaded"
            }
        case 7:
            // game over
            script = "Game over. Final score: " + ScoreScript()
        case 8:
            // hit
            script = text
        case 12:
            // ball
            b := " Ball "
            r := GetRand()
            if r < 0.125 {
                b += "high and inside"
            } else if r >= 0.125 && r < 0.25 {
                b += "high"
            } else if r >= 0.25 && r < 0.375 {
                b += "high and outside"
            } else if r >= 0.375 && r < 0.5 {
                b += "inside"
            } else if r >= 0.5 && r < 0.625 {
                b += "outside"
            } else if r >= 0.625 && r < 0.75 {
                b += "low and inside"
            } else if r >= 0.75 && r < 0.825 {
                b += "low"
            } else {
                b += "low and outside"
            }
            script = b + ". " + CountScript()
        case 13:
            // strike
            script = " " + text + ". " + CountScript()
        case 14:
            // out
            script = text
        case 15:
            // out #
            script = fmt.Sprintf("%d out", Inning.outs)
            if Inning.outs == 2 {
                script += "s"
            }
        case 16:
            // double play (whether successful or failed attempt)
            script = text
        case 17:
            // run(s) score(s)
            script = text + ". " + ScoreScript()
        case 18:
            // error (the text is in no way a reflection on which player caused the error, it's just a random position)
            tPos := Teams[0].batters[int(math.Floor(GetRand()*9))].pos
            if tPos == "DH" {
                tPos = "P"
            }
            script = "Error on " + tPos
    }
    PlayNum ++
    FullGameScript += script + "\n"   // strconv.Itoa(PlayNum) + ") " +
}

func ScoreScript() string {
    return fmt.Sprintf("%s %d, %s %d", Teams[0].short, Teams[0].score, Teams[1].short, Teams[1].score)
}

func CountScript() string {
    if Count.balls == 4 {
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
    if r < 0.2 {
        field = "left field"
    } else if r >= 0.2 && r < 0.4 {
        field = "left center"
    } else if r >= 0.4 && r < 0.6 {
        field = "center field"
    } else if r >= 0.6 && r < 0.8 {
        field = "right center"
    } else {
        field = "right field"
    }
    return field
}

func InningScript () string {
    var is string = ""
    if bti == 0 {
        is += "top"
    } else {
        is += "bottom"
    }
    is += fmt.Sprintf(" of the %d", InningNum)
    switch InningFrame {
        case 0:
            fallthrough
        case 1:
            is += "st"
        case 2:
            fallthrough
        case 3:
            is += "nd"
        case 4:
            fallthrough
        case 5:
            is += "rd"
        default:
            is += "th"
    }
    return is
}
