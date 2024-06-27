package cair

import "time"

var fakePlaylist []Item = []Item{
	Item{
		Name:     "Health insuranace",
		Duration: 30 * time.Second,
		// started 5 seconds ago
		ScheduledAt:  time.Now().UTC().Add(-5 * time.Second),
		ThirdPartyId: "C00000000",
	},
	Item{
		Name:     "Delicious beverage",
		Duration: 15 * time.Second,
		// current item duration minus its progress
		ScheduledAt:  time.Now().UTC().Add((30 - 5) * time.Second),
		ThirdPartyId: "C00000000",
	},
	Item{
		Name:     "Interesting Series",
		Duration: 30 * time.Second,
		// current item, minus progress, plus next item
		ScheduledAt:  time.Now().UTC().Add((30 - 5 + 15) * time.Second),
		ThirdPartyId: "P00000000",
	},
	Item{
		Name:     "The News",
		Duration: 15 * time.Minute,
		// current item, minus progress, plus next items
		ScheduledAt:  time.Now().UTC().Add((30 - 5 + 15 + 30) * time.Second),
		ThirdPartyId: "T00000000",
	},
}
