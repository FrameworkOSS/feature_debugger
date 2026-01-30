package debugger

import (
	"fmt"
	"os"
	"time"

	"github.com/FrameworkOSS/event"
	"github.com/FrameworkOSS/feature_commands/handler"
	"github.com/FrameworkOSS/portal"
)

type Debugger struct {
	p       *portal.Portal
	include []string
	exclude []string
}

func NewDebugger(p *portal.Portal) (dbg *Debugger) {
	dbg = new(Debugger)
	dbg.p = p
	return
}

func (dbg *Debugger) API() int {
	return 0
}

func (dbg *Debugger) ID() string {
	return "debugger"
}

func (dbg *Debugger) Name() string {
	return "Debugger"
}

func (dbg *Debugger) Authors() []string {
	return []string{"JoshuaDoes"}
}

func (dbg *Debugger) Description() string {
	return "Logs all incoming events to stdout, but errors to stderr."
}

func (dbg *Debugger) Version() string {
	return "v0.0.1"
}

func (dbg *Debugger) Open() error {
	//Hack: We should really use events to query the name and settings, avoiding the need for the debugger to attach a portal.
	portalID := dbg.p.ID()
	portalOpts := dbg.p.OptionsGet()

	//Hack: We should really use events to tell the portal that we are ready, once again avoiding native portal attachments.
	dbg.p.FeatureReady(dbg.ID(), true)

	fmt.Printf("Listening to %s\n%s\n\n", portalID, portalOpts.String())
	return nil
}

func (dbg *Debugger) Close() (errs []error, retry bool) {
	return
}

func (dbg *Debugger) Input(e *event.Event) error {
	dbg.debug(e)
	return nil
}

func (dbg *Debugger) Output() (*event.Event, error) {
	return nil, nil
}

func (dbg *Debugger) debug(e *event.Event) {
	if dbg.exclude != nil {
		for _, v := range dbg.exclude {
			if e.GetID() == v {
				return
			}
		}
	} else if dbg.include != nil {
		found := false
		for _, v := range dbg.include {
			if e.GetID() == v {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}
	p := []byte(DebugEvent(e))
	if e.GetID() == "error" {
		os.Stderr.Write(p)
	} else {
		os.Stdout.Write(p)
	}
}

func (dbg *Debugger) Include(events ...string) *Debugger {
	dbg.exclude = nil
	dbg.include = events
	return dbg
}

func (dbg *Debugger) Exclude(events ...string) *Debugger {
	dbg.exclude = events
	dbg.include = nil
	return dbg
}

// DebugEvent returns a string formatting of an event.
func DebugEvent(e *event.Event) (s string) {
	s += "<<<< portal:" + e.GetPortal() + " producer:" + e.GetProducer() + " >>>>"
	if epochMilli := e.GetEpochMilli(); epochMilli > 0 {
		t := time.UnixMilli(int64(epochMilli))
		s += fmt.Sprintf(" %s", t.Format(time.RFC3339Nano))
	}
	s += fmt.Sprintf("\nid:%s", event.Key(e.GetID()))
	if channel := e.GetChannel(); channel != "" {
		s += fmt.Sprintf(" (channel:%s)", channel)
	}
	if parts := e.GetParticipants(); len(parts) > 0 {
		s += fmt.Sprintf(" targets:%v", parts)
	}
	s += "\n"
	if offsets := e.GetOffsets(); len(offsets) > 0 {
		s += fmt.Sprintf("Offsets: %v\n", e.GetOffsets())
	}
	if data := e.GetData(); len(data) > 0 {
		s += "<<<\n"
		switch e.GetID() {
		case "call":
			call, args, err := handler.NewCommandArgsEvent(e)
			if err != nil {
				panic(err)
			}
			s += call
			for i := 0; i < len(args); i++ {
				arg := args[i]
				s += fmt.Sprintf("\n%s (%s): ", arg.GetID(), arg.GetType())
				switch arg.GetType() {
				case handler.CommandArgTypeNumber:
					s += fmt.Sprintf("%d", arg.GetValueNumber())
				case handler.CommandArgTypeString:
					s += arg.GetValueString()
				default:
					s += fmt.Sprintf("0x%X", arg.GetValueBytes())
				}
			}
		case event.EVENT_RESPONSE, event.EVENT_ERROR:
			s += string(data)
		default:
			s += fmt.Sprintf("0x%X\n===\n%s", data, string(data))
		}
		s += "\n>>>\n"
	}
	return s
}
