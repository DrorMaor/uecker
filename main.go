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
    "github.com/hegedustibor/htgo-tts"
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

var FullGameScript string = ""
var AudioFileName int = 0

type inning struct {
    num int
    TopBottom bool  // top True, bottom False
    outs int
    LeadOff bool    // for the script, to say if it's the leadoff batter
    // runners on bases:
    first bool
    second bool
    third bool
    // we need this T/F value if it's a walkoff win, so that will end the inning & game
    Walkoff bool
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
    batter int  // # in the lineup
    pitcher int // # in the pitchers list
    score int
    AtBatNum int
    Boxscore boxscore
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

type boxscore struct {
    inning []int
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
    Inning.num = 0
    Inning.TopBottom = false
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
    // this is the heading of thet boxscore
    box += "    "
    inns := len(Teams[0].Boxscore.inning)
    for i := 1; i <= inns; i++ {
        box += fmt.Sprintf("%d ", i)
    }
    box += " R H E \n"
    // now both teams' #s
    for _, team := range Teams {
        box += team.short + " "
        for i := 0; i < inns; i++ {
            box += fmt.Sprintf("%d ", team.Boxscore.inning[i])
        }
        box += fmt.Sprintf(" %d %d %d\n", team.score, team.Boxscore.H, team.Boxscore.E)
    }
    FullGameScript += box + "\n"
}

func StartInning() {
    Inning.LeadOff = true
    Inning.outs = 0
    Inning.first = false
    Inning.second = false
    Inning.third = false
    if !Inning.TopBottom {
        Inning.num ++
        bti = 0
    } else {
        bti = 1
    }
    Inning.TopBottom = !Inning.TopBottom
    Teams[bti].Boxscore.inning = append(Teams[bti].Boxscore.inning, 0)
    for {
        DoAtBat()
        if Inning.outs == 3 {
            EndInning()
            break
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
    GameScript(2, "")
    // first 8 innings, always start the next inning (frame)
    if Inning.num < 9 {
        StartInning()
    } else if Inning.num == 9 {
        DoNinthInning()
    } else {
        DoExtraInnings()
    }
}

func DoNinthInning() {
    // ------------------
    // 9th inning options
    // ------------------

    // (we flip the frame here, because it got flipped at the beginning of the inning)
    var TempFrame bool = !Inning.TopBottom

    // end of top of 9th, home team ahead, game over
    if Inning.num == 9 && !TempFrame && Teams[1].score > Teams[0].score {
        GameOver()
    }
    // end of top of 9th, home team NOT ahead, continue play
    if Inning.num == 9 && !TempFrame && Teams[1].score < Teams[0].score {
        StartInning()
    }
    // end of bottom of 9th, home team ahead, game over
    if Inning.num == 9 && TempFrame && Teams[1].score > Teams[0].score {
        GameOver()
    }
    // end of bottom of 9th, game tied, continue play
    if Inning.num == 9 && TempFrame && Teams[1].score == Teams[0].score {
        StartInning()
    }
    // end of bottom of 9th, away team ahead, game over
    if Inning.num == 9 && TempFrame && Teams[1].score < Teams[0].score {
        GameOver()
    }
}

func DoExtraInnings() {
    // ---------------------
    // extra innings options
    // ---------------------

    // (we flip the frame here, because it got flipped at the beginning of the inning)
    var TempFrame bool = !Inning.TopBottom

    // end of top of frame, always keep on playing
    if Inning.num > 9 && !TempFrame {
        StartInning()
    }
    // end of bottom of frame, game tied, continue playing
    if Inning.num > 9 && TempFrame && Teams[1].score == Teams[0].score {
        StartInning()
    }
    // end of bottom of frame, one team leads, game over
    if Inning.num > 9 && TempFrame && Teams[1].score != Teams[0].score {
        GameOver()
    }
}

func DoPitch() {
    if GetRand() < 0.75 {
        // ball or strike
        if GetRand() < 0.5 {
            // ball
            Count.balls ++
            GameScript(12, "")
            if Count.balls == 4 {
                // walk
                Count.balls = 0
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
                    GameScript(13, "Called string")
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
        if GetRand() < Teams[bti].batters[Teams[bti].AtBatNum].AVG {
            // he's on base
            // determine which hit type (param is # of bases in hit)
            r := GetRand()
            if r < 0.1 {
                DoHit(4, "Homerun to " + RandomField())
            } else if r >= 0.1 && r < 0.15 {
                // triple (can't use RandomField here, because a triple will rarely NOT be in right field)
                tText := "Triple to "
                if GetRand() < 0.5 {
                    tText += "right center"
                } else {
                    tText += "right field"
                }
                DoHit(3, tText)
            } else if r >= 0.15 && r < 0.33 {
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

    // most hits are out of the infield, so we assume them here
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

func TryDoublePlay(pos int) string {
    var dpText string = ""
    switch pos {
        case 1:
            switch BasesStatus() {
                case "false|false|false":
                case "true|false|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                    dpText = "1-4-3"
                case "false|true|false":
                case "false|false|true":
                case "true|true|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dpText = "1-4-3"
                case "true|false|true":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dpText = "1-4-3"
                case "false|true|true":
                case "true|true|true":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = false
                    dpText = "1-2-5"
            }
        case 2:
            // rare to have a catcher start a double play
        case 3:
            switch BasesStatus() {
                case "false|false|false":
                case "true|false|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                    dpText = "3-4-3"
                case "false|true|false":
                case "false|false|true":
                case "true|true|false":
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                    dpText = "3-4-3"
                case "true|false|true":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dpText = "3-4-3"
                case "false|true|true":
                case "true|true|true":
                    Inning.first = false
                    Inning.second = true
                    Inning.third = true
                    dpText = "3-4-3"
            }
        case 4:
            switch BasesStatus() {
                case "false|false|false":
                case "true|false|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                    dpText = "4-6-3"
                case "false|true|false":
                case "false|false|true":
                case "true|true|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dpText = "4-6-3"
                case "true|false|true":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dpText = "4-6-3"
                case "false|true|true":
                case "true|true|true":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = false
                    dpText = "4-2-5"
            }
        case 5:
            switch BasesStatus() {
                case "false|false|false":
                case "true|false|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                    dpText = "5-4-3"
                case "false|true|false":
                case "false|false|true":
                case "true|true|false":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = false
                    dpText = "5-4"
                case "true|false|true":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dpText = "5-4-3"
                case "false|true|true":
                case "true|true|true":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = false
                    dpText = "5-2"
            }
        case 6:
            switch BasesStatus() {
                case "false|false|false":
                case "true|false|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                    dpText = "6-4-3"
                case "false|true|false":
                case "false|false|true":
                case "true|true|false":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = false
                    dpText = "6-5"
                case "true|false|true":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dpText = "6-4-3"
                case "false|true|true":
                case "true|true|true":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = false
                    dpText = "6-2-5"
            }
    }
    return dpText
}

func IncrementScore(runs int, bases int) {
    // bases is used to determine if it's a walkoff
    if bases < 4 && Inning.num == 9 && Inning.TopBottom && (Teams[1].score - Teams[0].score < runs - 1) {
        Teams[bti].score ++
        GameScript(17, "1 run scores")
        Inning.Walkoff = true
        GameOver()
    } else {
        Teams[bti].score += runs
        rText := ""
        switch (runs) {
            case 1:
                rText = "1 run scores"
            default:
                rText = strconv.Itoa(runs) + " runs score"
        }
        GameScript(17, rText)
    }
    Teams[bti].Boxscore.inning[Inning.num-1] += runs
}

func AdvanceRunners(bases int, pos int) {
    // bases: # of bases of hit
    // pos: defensive position where ball was hit (1 based)

    var DoGameScript bool = true  // if it's a flyout and not a sac fly, then no runners advanced, so nothing to print
    switch bases {
        case -2: // error (assumed one base advance per runner, plus batter safe at first)
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
                    Inning.second = false
                    Inning.third = true
                case "false|false|true":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = false
                    IncrementScore(1, bases)
                case "true|true|false":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = true
                case "true|false|true":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = false
                    IncrementScore(1, bases)
                case "false|true|true":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = true
                    IncrementScore(1, bases)
                case "true|true|true":
                    IncrementScore(1, bases)
                    Inning.first = true
                    Inning.second = true
                    Inning.third = true
            }
        case -1: // out (sac fly)
            if pos >= 7 && Inning.third && Inning.outs < 3 {
                Inning.third = false  // the other 2 baserunners stay the same
                IncrementScore(1, bases)
            } else {
                DoGameScript = false
            }
        case 0: // walk
            switch BasesStatus() {
                case "false|false|false":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = false
                case "true|false|false":
                    fallthrough
                case "false|true|false":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = false
                case "false|false|true":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = true
                case "true|true|false":
                    fallthrough
                case "true|false|true":
                    fallthrough
                case "false|true|true":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = true
                case "true|true|true":
                    IncrementScore(1, bases)
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
                    if pos == 9 {
                        Inning.second = false
                        Inning.third = true
                    } else {
                        Inning.second = true
                        Inning.third = false
                    }
                case "false|true|false":
                    fallthrough
                case "false|false|true":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = false
                    IncrementScore(1, bases)
                case "true|true|false":
                    Inning.first = true
                    if pos >= 8 {
                        Inning.second = false
                        Inning.third = false
                        IncrementScore(1, bases)
                    } else {
                        Inning.second = false
                        Inning.third = true
                    }
                case "true|false|true":
                    Inning.first = true
                    Inning.third = false
                    IncrementScore(1, bases)
                    if pos == 9 {
                        Inning.third = true
                    } else {
                        Inning.second = true
                    }
                case "false|true|true":
                    Inning.first = true
                    Inning.second = false
                    if pos >= 8 {
                        IncrementScore(2, bases)
                        Inning.third = false
                    } else {
                        IncrementScore(1, bases)
                        Inning.third = true
                    }
                case "true|true|true":
                    Inning.first = true
                    Inning.third = true
                    if pos >= 8 {
                        IncrementScore(2, bases)
                        Inning.second = false
                    } else {
                        IncrementScore(1, bases)
                        Inning.second = true
                    }
            }
        case 2:
            switch BasesStatus() {
                case "false|false|false":
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "true|false|false":
                    fallthrough
                case "false|true|false":
                    fallthrough
                case "false|false|true":
                    IncrementScore(1, bases)
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "true|true|false":
                    fallthrough
                case "true|false|true":
                    fallthrough
                case "false|true|true":
                    IncrementScore(2, bases)
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "true|true|true":
                    IncrementScore(3, bases)
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
            }
        case 3:
            // will always clear the bases (and will put current batter on third)
            Inning.first = false
            Inning.second = false
            Inning.third = true
            switch BasesStatus() {
                case "true|false|false":
                    fallthrough
                case "false|true|false":
                    fallthrough
                case "false|false|true":
                    IncrementScore(1, bases)
                case "true|true|false":
                    fallthrough
                case "true|false|true":
                    fallthrough
                case "false|true|true":
                    IncrementScore(2, bases)
                case "true|true|true":
                    IncrementScore(3, bases)
            }
        case 4:
            // will always clear the bases
            Inning.first = false
            Inning.second = false
            Inning.third = false
            switch BasesStatus() {
                case "false|false|false":
                    IncrementScore(1, bases)
                case "true|false|false":
                    fallthrough
                case "false|true|false":
                    fallthrough
                case "false|false|true":
                    IncrementScore(2, bases)
                case "true|true|false":
                    fallthrough
                case "true|false|true":
                    fallthrough
                case "false|true|true":
                    IncrementScore(3, bases)
                case "true|true|true":
                    IncrementScore(4, bases)
            }
    }
    if DoGameScript {
        GameScript(6, "")
    }
}

func BasesStatus() string {
    return strconv.FormatBool(Inning.first) + "|" + strconv.FormatBool(Inning.second)  + "|" + strconv.FormatBool(Inning.third)
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
                dbText := TryDoublePlay(pos)
                if GetRand() < 0.85 && dbText != "" {
                    Inning.outs ++  // this will only be the EXTRA out
                    GameScript(16, dbText)
                }
            }
        }
    }
    Count.strikes = 0
    Count.balls = 0
    Inning.outs ++  // always increment the regular out
    AdvanceLineup()
    if Inning.outs == 3 || Inning.Walkoff {
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

func GetLineup(FileName string, Team team ) team {
    file, _ := os.Open(FileName)
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
            case "short":
                Team.short = line[1]
            case "WL":
                W, _ := strconv.ParseFloat(line[1], 64)
                L, _ := strconv.ParseFloat(line[2], 64)
                Team.pct = W / (W + L)
        	default:  // regular position players
        		Team.batters = append(Team.batters, batter{line[0], line[1], line[2], stat})
        }
	}
    return Team
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
                    script += fmt.Sprintf("%s, %s %s (AVG: %.3f)\n", batter.position, batter.FirstName, batter.LastName, batter.AVG)
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
            script += fmt.Sprintf(" for %s, %s %s %s", Teams[bti].short, Teams[bti].batters[Teams[bti].AtBatNum].position, Teams[bti].batters[Teams[bti].AtBatNum].FirstName, Teams[bti].batters[Teams[bti].AtBatNum].LastName)

            Inning.LeadOff = false
        case 6:
            // after runners advancing
            switch BasesStatus() {
                case "true|false|false":
                    script = "A runner at first base"
                case "false|true|false":
                    script = "A runner at second base"
                case "false|false|true":
                    script = "A runner at third base"
                case "true|true|false":
                    script = "Runners at first and second"
                case "true|false|true":
                    script = "Runners at first and third"
                case "false|true|true":
                    script = "Runners at second and third"
                case "true|true|true":
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
            // double play
            script = "Double play " + text
        case 17:
            // run(s) score(s)
            script = text + ". " + ScoreScript()
        case 18:
            // error (the text is in no way a reflection on which player caused the error, it's just a random position)
            tPos := Teams[0].batters[int(math.Floor(GetRand()*9)) + 1].position
            if tPos == "DH" {
                tPos = "P"
            }
            script = "Error on " + tPos
    }
    //fmt.Println(script)
    FullGameScript += script + "\n"
    // TTS(script)
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
    if Inning.TopBottom {
        is += "top"
    } else {
        is += "bottom"
    }
    is += fmt.Sprintf(" of the %d", Inning.num)
    switch Inning.num {
        case 1:
            is += "st"
        case 2:
            is += "nd"
        case 3:
            is += "rd"
        default:
            is += "th"
    }
    return is
}

func TTS(script string) {
    AudioFileName ++
    speech := htgotts.Speech{Folder: "audio", Language: "en"}
    speech.Speak(script)
}
