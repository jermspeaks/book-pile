package suggest

import (
	"fmt"
	"sort"

	"github.com/jermspeaks/book-pile/internal/store"
)

const ownedBonus = 3.0
const wantWeight = 10.0
const affinityWeight = 2.0

type ScoredBook struct {
	Book           store.Book
	Score          float64
	WantRating     int64
	AuthorAffinity int
	Owned          bool
	Reason         string
}

// RankNext scores unread books: want (primary), author affinity (read books by
// the same authors), and a small bonus for already-owned titles over coveting.
func RankNext(unread []store.Book, authorIDs map[int64][]int64, readCountByAuthor map[int64]int) []ScoredBook {
	out := make([]ScoredBook, 0, len(unread))
	for _, b := range unread {
		want := int64(0)
		if b.WantRating != nil {
			want = *b.WantRating
		}
		owned := b.OwnedPhysical != 0 || b.OwnedDigital != 0
		affinity := 0
		seen := map[int64]struct{}{}
		for _, aid := range authorIDs[b.ID] {
			if _, ok := seen[aid]; ok {
				continue
			}
			seen[aid] = struct{}{}
			affinity += readCountByAuthor[aid]
		}
		score := float64(want)*wantWeight + float64(affinity)*affinityWeight
		if owned {
			score += ownedBonus
		}
		out = append(out, ScoredBook{
			Book:           b,
			Score:          score,
			WantRating:     want,
			AuthorAffinity: affinity,
			Owned:          owned,
			Reason:         formatReason(want, affinity, owned),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		if out[i].WantRating != out[j].WantRating {
			return out[i].WantRating > out[j].WantRating
		}
		return out[i].Book.Title < out[j].Book.Title
	})
	return out
}

func formatReason(want int64, affinity int, owned bool) string {
	wantBit := "Unrated"
	if want > 0 {
		wantBit = fmt.Sprintf("Want %d", want)
	}
	ownedBit := "coveting"
	if owned {
		ownedBit = "already owned"
	}
	if affinity > 0 {
		noun := "book"
		if affinity != 1 {
			noun = "books"
		}
		return fmt.Sprintf("%s · %d other %s by this author · %s", wantBit, affinity, noun, ownedBit)
	}
	return wantBit + " · " + ownedBit
}
