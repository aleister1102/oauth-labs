package victims

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/cyllective/oauth-labs/victim/internal/browser"
	"github.com/cyllective/oauth-labs/victim/internal/config"
	"github.com/cyllective/oauth-labs/victim/internal/dto"
)

var (
	onceInit   = new(sync.Once)
	vmapMu     = new(sync.Mutex)
	vmap       = make(map[int]Victim, 0)
	allVictims = make([]Victim, 0)
)

func Init() {
	onceInit.Do(func() {
		cfg := config.Get()
		for rawVictimNum := range cfg.GetStringMap("victims") {
			victimNum, err := strconv.Atoi(rawVictimNum)
			if err != nil {
				log.Printf("warning: failed to load victim %s: %s", rawVictimNum, err)
				continue
			}
			v, err := buildVictim(victimNum)
			if err != nil {
				log.Printf("warning: failed to load victim %d: %s", victimNum, err)
				continue
			}
			vmap[victimNum] = v
			allVictims = append(allVictims, v)
		}

		slices.SortFunc(allVictims, func(a Victim, b Victim) int {
			if a.Number() < b.Number() {
				return -1
			} else if a.Number() > b.Number() {
				return 1
			}
			return 0
		})
	})
}

type VictimConfig struct {
	LabName   string
	ServerURL string `mapstructure:"server_url"`
	ClientURL string `mapstructure:"client_url"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	LabNumber int
}

type Victim interface {
	Number() int
	Name() string
	Config() VictimConfig
	Handle(browser *browser.Browser, url string, wait time.Duration) error
	CheckURL(string) error
}

func buildVictim(num int) (Victim, error) {
	cfg := config.Get()
	var vc VictimConfig
	if err := cfg.UnmarshalKey(fmt.Sprintf("victims.%d", num), &vc); err != nil {
		return nil, errors.New("unknown victim")
	}
	vc.LabNumber = num
	vc.LabName = fmt.Sprintf("%02d", num)

	switch num {
	case 2:
		return Victim02{vc}, nil
	case 3:
		return Victim03{vc}, nil
	default:
		return nil, errors.New("unknown victim")
	}
}

func All() []Victim {
	return allVictims
}

func Get(number int) (Victim, bool) {
	defer vmapMu.Unlock()
	vmapMu.Lock()

	v, ok := vmap[number]
	if !ok {
		return nil, false
	}
	return v, true
}

func Exists(number int) bool {
	_, ok := Get(number)
	return ok
}

func Handle(visitChan chan *dto.VisitRequest) {
	cfg := config.Get()
	browserConfig := &browser.Config{
		GlobalTimeout: time.Duration(cfg.GetInt("browser.globalTimeout")) * time.Second,
		BrowserBin:    cfg.GetString("browser.bin"),
	}

	for {
		req := <-visitChan
		v, ok := Get(req.LabNumber)
		if !ok {
			return
		}
		log.Printf("[victim%02d] running...", v.Number())
		b := browser.New(browserConfig)
		err := v.Handle(b, req.URL, time.Duration(3)*time.Second)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Printf("[victim%02d] timeout", v.Number())
			} else {
				log.Printf("[victim%02d] error: %s", v.Number(), err)
			}
			b.Close()
			continue
		}
		b.Close()
		log.Printf("[victim%02d] done.", v.Number())
	}
}
