package handler

import (
	"net/http"
	"strconv"
	"time"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// ページングの既定値と上限。apispec の default/minimum/maximum と揃える。
const (
	defaultLimit = 20
	maxLimit     = 100
)

// parseMatchFilter は履歴・集計で共通のクエリを読む。
//
// 空文字は「絞らない」として扱い、値が入っているときだけ検証する。
// 不正な値を黙って無視すると、絞ったつもりの結果が全件になって気付けないため。
func parseMatchFilter(r *http.Request, withPaging bool) (usecase.MatchFilter, error) {
	q := r.URL.Query()

	from, err := parseTime(q.Get("from"))
	if err != nil {
		return usecase.MatchFilter{}, err
	}

	to, err := parseTime(q.Get("to"))
	if err != nil {
		return usecase.MatchFilter{}, err
	}

	if from != nil && to != nil && to.Before(*from) {
		return usecase.MatchFilter{}, domain.ErrInvalidDateRange
	}

	mode := q.Get("mode")
	if mode != "" && !isKnownMode(mode) {
		return usecase.MatchFilter{}, domain.ErrInvalidMatchMode
	}

	outcome := q.Get("outcome")
	if outcome != "" && !isKnownOutcome(outcome) {
		return usecase.MatchFilter{}, domain.ErrInvalidOutcome
	}

	f := usecase.MatchFilter{From: from, To: to, Mode: mode, Outcome: outcome}

	if withPaging {
		f.Limit = parseBoundedInt(q.Get("limit"), defaultLimit, 1, maxLimit)
		f.Offset = parseBoundedInt(q.Get("offset"), 0, 0, 0)
	}

	return f, nil
}

// parseTime は RFC3339 と日付だけ (2026-08-20) の両方を受ける。
// フロントの日付ピッカーは日付だけを送ってくるため。
func parseTime(v string) (*time.Time, error) {
	if v == "" {
		return nil, nil
	}

	for _, layout := range []string{time.RFC3339, time.DateOnly} {
		if t, err := time.Parse(layout, v); err == nil {
			return &t, nil
		}
	}

	return nil, errInvalidQuery
}

// parseBoundedInt は範囲に収めた整数を返す。max が 0 なら上限なし。
// 範囲外はエラーにせず丸める。limit=1000 を 400 で突き返すより、
// 上限まで返すほうがクライアントにとって扱いやすい。
func parseBoundedInt(v string, fallback, minValue, maxValue int) int {
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	if n < minValue {
		return minValue
	}
	if maxValue > 0 && n > maxValue {
		return maxValue
	}

	return n
}

func isKnownMode(v string) bool {
	switch v {
	case domain.ModeRanked, domain.ModeCasual, domain.ModeAI, domain.ModeFriend:
		return true
	}

	return false
}

func isKnownOutcome(v string) bool {
	switch v {
	case domain.OutcomeWin, domain.OutcomeLoss, domain.OutcomeDraw:
		return true
	}

	return false
}
