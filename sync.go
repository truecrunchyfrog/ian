package ian

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
)

const CooldownJournalFilename string = ".cooldown-journal.toml"

type SyncEventType int

const (
	SyncEventPing SyncEventType = 1 << iota
	SyncEventCreate
	SyncEventUpdate
	SyncEventDelete
)

type SyncEvent struct {
	Type    SyncEventType
	Files   string
	Message string
}

type SyncCooldownInfo struct {
	Cooldowns map[string]time.Time
}

// Sync is called whenever changes are made to event(s), and calls any configured commands.
func (instance *Instance) Sync(eventInfo SyncEvent, ignoreCooldowns bool, stdout io.Writer) error {
	var cooldownJournal *SyncCooldownInfo
	var isJournalChanged bool

	for name, listener := range instance.Config.Sync.Listeners {
		if listener.Type == 0 || listener.Type&eventInfo.Type != 0 { // Type match
			ready := true

			if listener.Cooldown_ != 0 && !ignoreCooldowns {
				if cooldownJournal == nil {
					buf, err := os.ReadFile(filepath.Join(instance.Root, CooldownJournalFilename))
					if err != nil && !os.IsNotExist(err) {
						return err
					}
					if _, err := toml.Decode(string(buf), &cooldownJournal); err != nil {
						return err
					}
					if cooldownJournal.Cooldowns == nil {
						cooldownJournal.Cooldowns = map[string]time.Time{}
					}
				}

				// Don't need to check if map item exists with 'ok' because if it doesn't, lastChange will be 0 and it will work anyway.
				lastChange := cooldownJournal.Cooldowns[name]
				if now := time.Now(); lastChange.Add(listener.Cooldown_).Before(now) {
					// Cooldown gone
					cooldownJournal.Cooldowns[name] = now
					isJournalChanged = true
				} else {
					// Still in cooldown
					ready = false
				}
			}

			if ready {
				if stdout != nil {
					stdout.Write([]byte(fmt.Sprintf("\033[2m>>> BEGIN '%s'\033[22m\n", name)))
				}
				var cmd *exec.Cmd
				switch runtime.GOOS {
				case "windows":
					cmd = exec.Command(listener.Command)
				default:
					cmd = exec.Command("sh", "-c", listener.Command)
				}
				cmd.Stdout = stdout
				buf := new(bytes.Buffer)
				cmd.Stderr = buf
				absRoot, err := filepath.Abs(instance.Root)
				if err != nil {
					return err
				}
				cmd.Dir = absRoot
				cmd.Env = append(os.Environ(),
					"MESSAGE="+eventInfo.Message,
					"FILES="+eventInfo.Files,
					"TYPE="+fmt.Sprint(eventInfo.Type),
				)
				if err := cmd.Run(); err != nil {
					log.Printf("warning: sync listener command '%s' exited unsuccessfully (%s). stderr:\n%s", name, err, buf.String())
				}
				if stdout != nil {
					stdout.Write([]byte(fmt.Sprintf("\033[2m<<< END   '%s'\033[22m\n\n", name)))
				}
			}
		}
	}

	if isJournalChanged {
		buf := new(bytes.Buffer)
		if err := toml.NewEncoder(buf).Encode(cooldownJournal); err != nil {
			return err
		}
		os.WriteFile(filepath.Join(instance.Root, CooldownJournalFilename), buf.Bytes(), 0644)
	}

	return nil
}
