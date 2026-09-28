package suggest

import (
	"testing"

	"github.com/jermspeaks/book-pile/internal/store"
)

func TestRankNextWantThenAffinityThenOwned(t *testing.T) {
	want5 := int64(5)
	want3 := int64(3)
	unread := []store.Book{
		{ID: 1, Title: "Low want owned", OwnedDigital: 1, WantRating: &want3},
		{ID: 2, Title: "High want covet", WantRating: &want5},
		{ID: 3, Title: "High want owned same author", OwnedDigital: 1, WantRating: &want5},
		{ID: 4, Title: "Unrated owned", OwnedPhysical: 1},
	}
	authorIDs := map[int64][]int64{
		1: {10},
		2: {20},
		3: {10},
		4: {30},
	}
	readCount := map[int64]int{10: 4}

	got := RankNext(unread, authorIDs, readCount)
	if len(got) != 4 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Book.ID != 3 {
		t.Fatalf("expected high-want + affinity first, got id=%d score=%.1f", got[0].Book.ID, got[0].Score)
	}
	if got[1].Book.ID != 2 {
		t.Fatalf("expected high-want covet second, got id=%d", got[1].Book.ID)
	}
	if got[2].Book.ID != 1 {
		t.Fatalf("expected lower want third, got id=%d", got[2].Book.ID)
	}
	if got[3].Book.ID != 4 {
		t.Fatalf("expected unrated last, got id=%d", got[3].Book.ID)
	}
	if got[0].Score <= got[1].Score {
		t.Fatalf("owned affinity book should outrank covet with same want: %.1f vs %.1f", got[0].Score, got[1].Score)
	}
}
