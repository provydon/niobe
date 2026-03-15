package proxy

import "testing"

func TestConfirmationRequiredBlocksInitialToolCallWithoutTranscript(t *testing.T) {
	state := &sessionState{}

	if !confirmationRequired(state, &guardedTurn{}) {
		t.Fatal("expected the first action tool call to require confirmation")
	}
}

func TestConfirmationRequiredAllowsConfirmedTurnByTranscript(t *testing.T) {
	state := &sessionState{}
	state.requestConfirmation()

	turn := &guardedTurn{}
	turn.userTranscript.WriteString("yes, go ahead")

	if confirmationRequired(state, turn) {
		t.Fatal("expected explicit confirmation text to unlock the action")
	}
}

func TestConfirmationRequiredAllowsNextTurnWhenTranscriptLags(t *testing.T) {
	state := &sessionState{}
	state.requestConfirmation()
	state.markUserTurnComplete()

	if confirmationRequired(state, &guardedTurn{}) {
		t.Fatal("expected the next completed user turn to unlock the action even before transcription arrives")
	}
}

func TestConfirmationRequiredAllowsActiveNextTurnBeforeTurnComplete(t *testing.T) {
	state := &sessionState{}
	state.requestConfirmation()
	state.markUserTurnStarted()

	if confirmationRequired(state, &guardedTurn{}) {
		t.Fatal("expected an active next user turn to unlock the action before audioStreamEnd arrives")
	}
}

func TestShouldBlockOptimisticClaimAfterBlockedToolCall(t *testing.T) {
	turn := &guardedTurn{}
	turn.transcript.WriteString("I've placed the order for you.")

	if !shouldBlockOptimisticClaim(turn) {
		t.Fatal("expected optimistic completion claim to be blocked without a successful Laravel action")
	}
}

func TestShouldAllowOptimisticClaimAfterSuccessfulAction(t *testing.T) {
	turn := &guardedTurn{actionSucceeded: true}
	turn.transcript.WriteString("I've placed the order for you.")

	if shouldBlockOptimisticClaim(turn) {
		t.Fatal("expected completion claim to be allowed after a successful Laravel action")
	}
}

func TestSynchronousActionSucceededRequiresOkTrue(t *testing.T) {
	result := map[string]any{
		"ok": true,
	}

	if !synchronousActionSucceeded(result) {
		t.Fatal("expected ok=true to count as a successful Laravel response")
	}
}

func TestSynchronousActionSucceededRejectsOkFalse(t *testing.T) {
	result := map[string]any{
		"ok":      false,
		"message": "Action failed.",
	}

	if synchronousActionSucceeded(result) {
		t.Fatal("expected ok=false to be treated as an unsuccessful Laravel response")
	}
}

func TestNormalizeEventStatusPromotesQueuedOkResponseToSuccess(t *testing.T) {
	result := map[string]any{
		"ok": true,
	}

	if got := normalizeEventStatus("queued", result); got != "success" {
		t.Fatalf("expected queued ok response to be normalized to success, got %q", got)
	}
}
