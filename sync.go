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
	"strings"
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
	Files   []string
	Message string
}

type SyncCooldownInfo struct {
	Cooldowns map[string]time.Time
}

// Sync is called whenever changes are made to event(s), with the changes occuring in action, and calls any configured commands.
func (instance *Instance) Sync(action func() error, eventInfo SyncEvent, ignoreCooldowns bool, stdouterr io.Writer) error {
	hooks := map[string]Hook{}

	var cooldownJournal *SyncCooldownInfo
	var isJournalChanged bool

	for name, hook := range instance.Config.Hooks {
		if hook.Type == 0 || hook.Type&eventInfo.Type != 0 { // Type match
			ready := true

			if hook.Cooldown_ != 0 && !ignoreCooldowns {
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
				if now := time.Now(); lastChange.Add(hook.Cooldown_).Before(now) {
					// Cooldown gone
					cooldownJournal.Cooldowns[name] = now
					isJournalChanged = true
				} else {
					// Still in cooldown
					ready = false
				}
			}

			if ready {
				hooks[name] = hook
			}
		}
	}

	// PRE

	for name, hook := range hooks {
		if hook.PreCommand != "" {
			if stdouterr != nil {
				stdouterr.Write([]byte(fmt.Sprintf("\033[2m=== RUN  hook '%s' PRE-command\033[22m\n", name)))
			}

			err := runHookCommand(eventInfo, hook.PreCommand, instance.Root, stdouterr)

			if err != nil {
				log.Printf("warning: sync hook command '%s' exited unsuccessfully (%s).\n", name, err)
			}

			if stdouterr != nil {
				stdouterr.Write([]byte(fmt.Sprintf("\033[2m=== DONE hook '%s' PRE-command\033[22m\n", name)))
			}
		}
	}

	if stdouterr != nil {
		stdouterr.Write([]byte("\n\033[2m=== MODIFYING EVENTS\033[0m\n\n"))
	}

	if err := action(); err != nil {
		return err
	}

	// POST

	for name, hook := range hooks {
		if hook.PostCommand != "" {
			if stdouterr != nil {
				stdouterr.Write([]byte(fmt.Sprintf("\033[2m=== RUN  hook '%s' POST-command\033[22m\n", name)))
			}

			err := runHookCommand(eventInfo, hook.PostCommand, instance.Root, stdouterr)

			if err != nil {
				log.Printf("warning: sync hook command '%s' exited unsuccessfully (%s).\n", name, err)
			}

			if stdouterr != nil {
				stdouterr.Write([]byte(fmt.Sprintf("\033[2m=== DONE hook '%s' POST-command\033[22m\n", name)))
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

func runHookCommand(eventInfo SyncEvent, command string, workingDir string, stdouterr io.Writer) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command(command)
	default:
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Stdout = stdouterr
	cmd.Stderr = stdouterr

	absRoot, err := filepath.Abs(workingDir)
	if err != nil {
		return err
	}

	cmd.Dir = absRoot
	cmd.Env = append(os.Environ(),
		"MESSAGE="+eventInfo.Message,
		"FILES="+strings.Join(eventInfo.Files, " "),
		"TYPE="+fmt.Sprint(eventInfo.Type),
	)

	return cmd.Run()
}
