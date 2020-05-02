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
    short string  // 3 letter short team (such as MIL)
    pct float64   // winning percentage: W/(W+L)
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
    Inning.num = 0
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
    if !Inning.TopBottom {
        Inning.num ++
        bti = 0
    } else {
        bti = 1
    }
    Inning.TopBottom = !Inning.TopBottom

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
    // all this occurs at the end of an inning, BEFORE we increment the inning # and frame
    // (which we do at the beginning of StartInning)
    GameScript(2)
    // first 8 innings, always start the next inning (frame)
    if Inning.num < 9  {
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
    if GetRand() < .75 {
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
            if ! (Count.strikes == 2 && s >= .667) {
                // only add strike if it's not Strike 2 now and it's not a foul ball
                Count.strikes ++
                if s < .333 {
                    GameScript(13)
                } else if s >= .333 && s < .667 {
                    GameScript(14)
                } else {
                    GameScript(15)
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
            if r < .1 {
                GameScript(8)
                DoHit(4)
            } else if r >= .1 && r < .15 {
                GameScript(9)
                DoHit(3)
            } else if r >= .15 && r < .33 {
                GameScript(10)
                DoHit(2)
            } else {
                GameScript(11)
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
    DoAtBat()
}

func AdvanceLineup() {
    Teams[bti].AtBatNum ++
    if Teams[bti].AtBatNum % 9 == 0 {
        Teams[bti].AtBatNum = 0
    }
}

func TryDoublePlay(pos float64) bool {
    var dp bool = false
    switch pos {
        case 1:
            switch BasesStatus() {
                case "false|false|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                    dp = false
                case "true|false|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                    dp = true
                case "false|true|false":
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                    dp = false
                case "false|false|true":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dp = false
                case "true|true|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dp = true
                case "true|false|true":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dp = true
                case "false|true|true":
                    Inning.first = false
                    Inning.second = true
                    Inning.third = true
                    dp = false
                case "true|true|true":
                    Inning.first = true
                    Inning.second = true
                    Inning.third = false
                    dp = true
            }
        case 2:
            dp = false
        case 3:
            switch BasesStatus() {
                case "false|false|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                    dp = false
                case "true|false|false":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                    dp = true
                case "false|true|false":
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                    dp = false
                case "false|false|true":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dp = false
                case "true|true|false":
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                    dp = true
                case "true|false|true":
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                    dp = true
                case "false|true|true":
                    Inning.first = false
                    Inning.second = true
                    Inning.third = true
                    dp = false
                case "true|true|true":
                    Inning.first = false
                    Inning.second = true
                    Inning.third = true
                    dp = true
            }
    }
    return dp
}

func AdvanceRunners(bases int, pos float64) {
    // bases: # of bases of hit
    // pos: defensive position where ball was hit

    var DoGameScript bool = true  // if it's a flyout and not a sac fly, then no runners advanced, so nothing to print
    switch bases {
        case -1: // out (sac fly)
            if pos >= 6 && Inning.third && Inning.outs < 3 {
                Inning.third = false  // the other 2 baserunners stay the same
                Teams[bti].score ++
                GameScript(50)
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
                    GameScript(50)
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
                    GameScript(50)
                case "false|false|true":
                    Inning.first = true
                    Inning.second = false
                    Inning.third = false
                    Teams[bti].score ++
                    GameScript(50)
                case "true|true|false":
                    Inning.first = true
                    if pos == 7 || pos == 8 {
                        Inning.second = false
                        Inning.third = false
                        Teams[bti].score ++
                        GameScript(50)
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
                    GameScript(50)
                case "false|true|true":
                    if pos == 7 || pos == 8 {
                        Teams[bti].score += 2
                        GameScript(51)
                        Inning.first = true
                        Inning.second = false
                        Inning.third = false
                    } else {
                        Teams[bti].score ++
                        GameScript(50)
                        Inning.first = true
                        Inning.second = false
                        Inning.third = true
                    }
                case "true|true|true":
                    if pos == 7 || pos == 8 {
                        Teams[bti].score += 2
                        GameScript(51)
                        Inning.first = true
                        Inning.second = false
                        Inning.third = false
                    } else {
                        Teams[bti].score ++
                        GameScript(50)
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
                    GameScript(50)
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "false|true|false":
                    Teams[bti].score ++
                    GameScript(50)
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "false|false|true":
                    Teams[bti].score ++
                    GameScript(50)
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "true|true|false":
                    Teams[bti].score += 2
                    GameScript(51)
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "true|false|true":
                    Teams[bti].score += 2
                    GameScript(51)
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "false|true|true":
                    Teams[bti].score += 2
                    GameScript(51)
                    Inning.first = false
                    Inning.second = true
                    Inning.third = false
                case "true|true|true":
                    Teams[bti].score += 3
                    GameScript(52)
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
                    GameScript(50)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "false|true|false":
                    Teams[bti].score ++
                    GameScript(50)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "false|false|true":
                    Teams[bti].score ++
                    GameScript(50)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "true|true|false":
                    Teams[bti].score += 2
                    GameScript(51)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "true|false|true":
                    Teams[bti].score += 2
                    GameScript(51)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "false|true|true":
                    Teams[bti].score += 2
                    GameScript(51)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
                case "true|true|true":
                    Teams[bti].score += 3
                    GameScript(52)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = true
            }
        case 4:
            switch BasesStatus() {
                case "false|false|false":
                    Teams[bti].score ++
                    GameScript(50)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "true|false|false":
                    Teams[bti].score += 2
                    GameScript(51)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "false|true|false":
                    Teams[bti].score += 2
                    GameScript(51)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "false|false|true":
                    Teams[bti].score += 2
                    GameScript(51)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "true|true|false":
                    Teams[bti].score += 3
                    GameScript(52)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "true|false|true":
                    Teams[bti].score += 3
                    GameScript(52)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "false|true|true":
                    Teams[bti].score += 3
                    GameScript(52)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
                case "true|true|true":
                    Teams[bti].score += 4
                    GameScript(53)
                    Inning.first = false
                    Inning.second = false
                    Inning.third = false
            }
    }
    if DoGameScript {
        GameScript(6)
    }
}

func BasesStatus() string {
    return strconv.FormatBool(Inning.first) + "|" + strconv.FormatBool(Inning.second)  + "|" + strconv.FormatBool(Inning.third)
}

func DoOut(strikeout bool) {
    if !strikeout {
        pos := math.Floor(GetRand()*9) +1
        GameScript(int(pos + 30))
        if pos >= 7 {
            AdvanceRunners(-1, pos)
        } else {
            if Inning.outs < 2 {
                if GetRand() < .85 && TryDoublePlay(pos) {
                    Inning.outs ++  // this will only be the EXTRA out
                    GameScript(55)
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
        GameScript(40)
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
            case "short":
                Team.short = line[1]
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
            // print the lineups
            for _, team := range Teams {
                script += fmt.Sprintf("%s's lineup:\n", team.city)
                for _, batter := range team.batters {
                    script += fmt.Sprintf("%s, %s %s\n", batter.position, batter.FirstName, batter.LastName)
                }
                script += "\n"
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
            // home run
            script = "Homerun to " + RandomField()
        case 9:
            // triple (can't use RandomField here, because a triple will rarely not be in right field)
            script = "Triple to "
            if GetRand() < .5 {
                script += "right center"
            } else {
                script += "right field"
            }
        case 10:
            // double
            script = "Double to " + RandomField()
        case 11:
            // single
            script = "Single to " + RandomField()
        case 12:
            // ball
            b := "Ball "
            r := GetRand()
            if r < .125 {
                b += "high and inside"
            } else if r >= .125 && r < .25 {
                b += "high"
            } else if r >= .25 && r < .375 {
                b += "high and outside"
            } else if r >= .375 && r < .5 {
                b += "inside"
            } else if r >= .5 && r < .625 {
                b += "outside"
            } else if r >= .625 && r < .75 {
                b += "low and inside"
            } else if r >= .75 && r < .825 {
                b += "low"
            } else {
                b += "low and outside"
            }
            script = b + ". " + CountScript()
        case 13:
            // swinging strike
            script = "Swing and a miss. " + CountScript()
        case 14:
            // called strike
            script = "Called strike. " + CountScript()
        case 15:
            // foul ball
            script = "Foul ball. " + CountScript()

        // these are all outs
        case 31:
            script = "Groundout to the pitcher."
        case 32:
            script = "Popout to the catcher."
        case 33:
            script = "Groundout to first."
        case 34:
            script = "Groundout to second."
        case 35:
            script = "Groundout to third."
        case 36:
            script = "Groundout to the shortstop."
        case 37:
            script = "Flyout to left."
        case 38:
            script = "Flyout to center."
        case 39:
            script = "Flyout to right."
        case 40:
            script = fmt.Sprintf("%d out", Inning.outs)
            if Inning.outs == 2 {
                script += "s"
            }

        // run(s) score(s)
        case 50:
            script = "1 run scores. " + ScoreScript()
        case 51:
            script = "2 runs score. " + ScoreScript()
        case 52:
            script = "3 runs score. " + ScoreScript()
        case 53:
            script = "4 runs score. " + ScoreScript()

        case 55:
            script = "Double play"
    }
    fmt.Println(script)
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
    if r < .2 {
        field = "left field"
    } else if r >= .2 && r < .4 {
        field = "left center"
    } else if r >= .4 && r < .6 {
        field = "center field"
    } else if r >= .6 && r < .8 {
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
