package checkedexamples

import (
	verity_reporting "github.com/verity-bdd/verity-bdd/verity_reporting"
)

func reportCustomStep(reporter verity_reporting.Reporter, err error) {
	tracker := verity_reporting.NewActivityTracker(reporter, "exports diagnostics")
	tracker.Start()
	tracker.Finish(err, verity_reporting.Attachment{
		Name:        "diagnostics.json",
		ContentType: "application/json",
		Content:     []byte(`{"status":"captured"}`),
	})
}
