package game

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func newTestFixture() *Game {
	return newGame([]PlayerID{"player-1", "player-2"}, PendingChoice{
		ID:      "fixture-opening-choice",
		Player:  "player-1",
		Options: []string{"left", "right"},
	})
}

func TestWalkingSkeleton(t *testing.T) {
	game := newTestFixture()
	view := game.PlayerView("player-1")
	if view.Revision != 1 || view.PendingChoice == nil {
		t.Fatalf("initial PlayerView = %#v, want pending choice at revision 1", view)
	}

	before := game.StateHash()
	if err := game.SubmitChoice("player-1", SubmitChoice{Revision: 0, Handle: view.PendingChoice.Handle, Option: "left"}); err == nil {
		t.Fatal("stale revision was accepted")
	}
	if game.StateHash() != before {
		t.Fatal("stale revision changed state")
	}
	if err := game.SubmitChoice("player-2", SubmitChoice{Revision: view.Revision, Handle: view.PendingChoice.Handle, Option: "left"}); err == nil {
		t.Fatal("cross-player handle was accepted")
	}
	if game.StateHash() != before {
		t.Fatal("cross-player handle changed state")
	}
	if err := game.SubmitChoice("player-1", SubmitChoice{Revision: view.Revision, Handle: view.PendingChoice.Handle, Option: "invalid"}); err == nil {
		t.Fatal("illegal option was accepted")
	}
	if game.StateHash() != before {
		t.Fatal("illegal option changed state")
	}

	if err := game.SubmitChoice("player-1", SubmitChoice{Revision: view.Revision, Handle: view.PendingChoice.Handle, Option: "left"}); err != nil {
		t.Fatalf("submit choice: %v", err)
	}
	if err := game.Concede("player-2"); err != nil {
		t.Fatalf("concede: %v", err)
	}
	if !game.PlayerView("player-1").Finished {
		t.Fatal("game did not finish after concession")
	}

	replay := game.Replay()
	if err := replay.Verify(newTestFixture()); err != nil {
		t.Fatalf("replay verification: %v", err)
	}
}

func TestFixtureCLISmoke(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestFixtureCLIHelper", "--")
	cmd.Env = append(os.Environ(), "GO_TCG_FIXTURE_CLI=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := stdin.Write([]byte("{\"player\":\"player-1\",\"action\":\"choice\",\"revision\":1,\"handle\":\"choice-player-1-r1\",\"option\":\"left\"}\n{\"player\":\"player-2\",\"action\":\"concede\"}\n")); err != nil {
		t.Fatal(err)
	}
	if err := stdin.Close(); err != nil {
		t.Fatal(err)
	}

	var lines []cliResponse
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		var line cliResponse
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			t.Fatal(err)
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 || lines[0].View == nil || !lines[2].ReplayValid || lines[2].Replay == nil {
		t.Fatalf("CLI responses = %#v, want final verified replay", lines)
	}
	if err := lines[2].Replay.Verify(newTestFixture()); err != nil {
		t.Fatalf("CLI emitted unverifiable replay: %v", err)
	}
}

type cliRequest struct {
	Player   PlayerID `json:"player"`
	Action   string   `json:"action"`
	Revision uint64   `json:"revision"`
	Handle   string   `json:"handle"`
	Option   string   `json:"option"`
}

type cliResponse struct {
	Error       string      `json:"error,omitempty"`
	ReplayValid bool        `json:"replay_valid,omitempty"`
	Replay      *Replay     `json:"replay,omitempty"`
	View        *PlayerView `json:"view,omitempty"`
}

func TestFixtureCLIHelper(t *testing.T) {
	if os.Getenv("GO_TCG_FIXTURE_CLI") != "1" {
		return
	}
	game := newTestFixture()
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	_ = encoder.Encode(cliResponse{View: ptr(game.PlayerView("player-1"))})
	for scanner.Scan() {
		var request cliRequest
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			_ = encoder.Encode(cliResponse{Error: err.Error()})
			continue
		}
		var err error
		switch request.Action {
		case "choice":
			err = game.SubmitChoice(request.Player, SubmitChoice{Revision: request.Revision, Handle: request.Handle, Option: request.Option})
		case "concede":
			err = game.Concede(request.Player)
		default:
			err = errInvalidInput
		}
		if err != nil {
			_ = encoder.Encode(cliResponse{Error: err.Error()})
			continue
		}
		response := cliResponse{}
		if game.PlayerView("player-1").Finished {
			replay := game.Replay()
			response.Replay = &replay
			response.ReplayValid = replay.Verify(newTestFixture()) == nil
		}
		_ = encoder.Encode(response)
	}
	if err := scanner.Err(); err != nil && !strings.Contains(err.Error(), "closed") {
		t.Fatal(err)
	}
	os.Exit(0)
}

func ptr[T any](value T) *T { return &value }
