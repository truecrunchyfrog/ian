package ian

import (
	"testing"
	"time"
)

func TestMigrateToThenFromIcal(t *testing.T) {
	now := time.Now().In(time.UTC).Truncate(time.Second)

	props := EventProperties{
		Uid:         GenerateUid(),
		Summary:     "summary",
		Description: "description",
		Location:    "location",
		Url:         "url",
		Start:       now,
		End:         now.Add(5 * time.Hour),
		Recurrence: Recurrence{
			RRule: "FREQ=DAILY;INTERVAL=3;COUNT=12",
		},
		Created:  now,
		Modified: now,
	}

	events := []Event{
		{
			Props: props,
		},
	}

	ical := ToIcal(events, "")
	native, err := FromIcal(ical)
	if err != nil {
		t.Error(err)
	}

	if native[0] != props {
		t.Errorf("migration to ical and back failed:\n\ngot:  %+v\nwant: %+v", native[0], props)
	}
}
